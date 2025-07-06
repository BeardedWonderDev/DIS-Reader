import java.io.*;
import java.net.*;
import java.sql.*;
import java.util.*;
import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import javax.sql.DataSource;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.core.type.TypeReference;

public class JDBCRunner {
    private static DataSource dataSource;
    private static final ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        // Disable GUI-based sign-on prompts
        System.setProperty("java.awt.headless", "true");
        // Disable IBM Toolbox GUI dialogs via system property
        System.setProperty("com.ibm.as400.access.guiAvailable", "false");

        if (args.length < 1) {
            System.err.println("Usage: java -jar <jar> <port>");
            System.exit(1);
        }

        int port = Integer.parseInt(args[0]);

        ServerSocket serverSocket = new ServerSocket(port);
        System.out.println(mapper.writeValueAsString(Map.of("status","ok","message","Server started on port "+port)));

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                if (dataSource instanceof HikariDataSource) {
                    ((HikariDataSource)dataSource).close();
                    dataSource = null;
                }
                serverSocket.close();
                System.out.println(mapper.writeValueAsString(Map.of("status","ok","message","Server stopped")));
            } catch (IOException e) {
                // ignore
            }
        }));

        while (!serverSocket.isClosed()) {
            try {
                Socket clientSocket = serverSocket.accept();
                new Thread(() -> handleClient(clientSocket)).start();
            } catch (SocketException se) {
                // Server socket closed, exit loop
                break;
            }
        }
    }

    private static void handleClient(Socket socket) {
        try {
            try {
                socket.setSoTimeout(30000);
            } catch (SocketException e) {
                // ignore
            }
            BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), "UTF-8"));
            BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), "UTF-8"));
            String line;
            while ((line = reader.readLine()) != null) {
                Map<String, String> cmdMap;
                try {
                    cmdMap = mapper.readValue(line, new TypeReference<Map<String, String>>() {});
                } catch (Exception e) {
                    writeLine(writer, mapper.writeValueAsString(Map.of("status","error","message","Invalid JSON: "+e.getMessage())));
                    break;
                }
                if (!cmdMap.containsKey("cmd")) {
                    writeLine(writer, mapper.writeValueAsString(Map.of("status","error","message","Missing 'cmd' field")));
                    break;
                }

                String requestId = cmdMap.get("requestId");
                String cmd = cmdMap.get("cmd").toLowerCase();
                switch (cmd) {
                    case "ping":
                        // Health check: ensure Java server and database connection are alive
                        if (dataSource == null) {
                            writeLine(writer, mapper.writeValueAsString(Map.of(
                                "status","error",
                                "message","Not connected to DB",
                                "requestId", requestId
                            )));
                        } else {
                            try (Connection testConn = dataSource.getConnection()) {
                                writeLine(writer, mapper.writeValueAsString(Map.of(
                                    "status","ok",
                                    "message","pong",
                                    "requestId", requestId
                                )));
                            } catch (Exception ex) {
                                writeLine(writer, mapper.writeValueAsString(Map.of(
                                    "status","error",
                                    "message","DB ping failed: " + ex.getMessage(),
                                    "requestId", requestId
                                )));
                            }
                        }
                        break;
                    case "connect":
                        handleConnect(writer, cmdMap);
                        break;
                    case "disconnect":
                        handleDisconnect(writer, requestId);
                        socket.close();
                        return;
                    case "query":
                        if (!cmdMap.containsKey("sql")) {
                            writeLine(writer, mapper.writeValueAsString(Map.of("status","error","message","Missing 'sql' field in query command","requestId", requestId)));
                        } else {
                            handleQuery(writer, cmdMap.get("sql"), socket, cmdMap);
                        }
                        break;
                    default:
                        writeLine(writer, mapper.writeValueAsString(Map.of("status","error","message","Unknown command: " + cmd,"requestId", requestId)));
                        break;
                }
            }
        } catch (IOException e) {
            // Client disconnected or error occurred
        } finally {
            // Do NOT close the shared JDBC connection here; only close the client socket.
            try {
                socket.close();
            } catch (IOException ignore) {}
        }
    }

    private static void handleConnect(BufferedWriter writer, Map<String,String> cmdMap) throws IOException {
        String requestId = cmdMap.get("requestId");
        String host = cmdMap.get("host");
        String user = cmdMap.get("user");
        String pass = cmdMap.get("pass");
        String url = "jdbc:as400://" + host + ";naming=system";

        try {
            if (dataSource instanceof HikariDataSource) {
                ((HikariDataSource)dataSource).close();
                dataSource = null;
            }
            HikariConfig poolConfig = new HikariConfig();
            poolConfig.setJdbcUrl(url);
            poolConfig.setUsername(user);
            poolConfig.setPassword(pass);
            // Optional pool tuning
            poolConfig.addDataSourceProperty("maximumPoolSize", "10");
            poolConfig.setMinimumIdle(2);
            poolConfig.setIdleTimeout(300000);
            poolConfig.setConnectionTestQuery("SELECT 1 FROM SYSIBM.SYSDUMMY1");
            dataSource = new HikariDataSource(poolConfig);

            try (Connection testConn = dataSource.getConnection()) {
                writeLine(writer, mapper.writeValueAsString(Map.of(
                    "status","ok",
                    "message","Connection pool ready",
                    "requestId", requestId
                )));
            }
        } catch (Exception ex) {
            writeLine(writer, mapper.writeValueAsString(Map.of(
                "status","error",
                "message","Pool init failed: " + ex.getMessage(),
                "requestId", requestId
            )));
            dataSource = null;
        }
    }

    private static void handleDisconnect(BufferedWriter writer, String requestId) throws IOException {
        if (dataSource instanceof HikariDataSource) {
            ((HikariDataSource)dataSource).close();
            dataSource = null;
            writeLine(writer, mapper.writeValueAsString(Map.of(
                "status","ok",
                "message","Pool closed",
                "requestId", requestId
            )));
            return;
        }
        writeLine(writer, mapper.writeValueAsString(Map.of(
            "status","ok",
            "message","No active pool to close",
            "requestId", requestId
        )));
    }

    private static void handleQuery(BufferedWriter writer, String sql, Socket socket, Map<String, String> cmdMap) throws IOException {
        String requestId = cmdMap.get("requestId");
        boolean includeSrc = "true".equalsIgnoreCase(cmdMap.get("includeSrc"));
        if (dataSource == null) {
            writeLine(writer, mapper.writeValueAsString(Map.of("status","error","message","Not connected","requestId", requestId)));
            writer.flush();
            return;
        }
        try (Connection conn = dataSource.getConnection();
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {

            ResultSetMetaData md = rs.getMetaData();
            int colCount = md.getColumnCount();

            while (rs.next()) {
                Map<String, String> row = new LinkedHashMap<>();
                for (int i = 1; i <= colCount; i++) {
                    String val = rs.getString(i);
                    row.put(md.getColumnName(i), val != null ? val : "");
                }
                if (includeSrc) {
                    row.put("src_table", extractTableName(sql));
                }
                row.put("requestId", requestId);
                writeLine(writer, mapper.writeValueAsString(row));
            }
            writeLine(writer, mapper.writeValueAsString(Map.of(
                "status","ok",
                "message","Query completed",
                "requestId", requestId
            )));
            writeLine(writer, mapper.writeValueAsString(Map.of(
                "status","done",
                "requestId", requestId
            )));
            writer.flush();
            return;
        } catch (Exception ex) {
            writeLine(writer, mapper.writeValueAsString(Map.of(
                "status","error",
                "message","Query failed: " + ex.getMessage(),
                "requestId", requestId
            )));
            writer.flush();
            return;
        }
    }

    private static void writeLine(BufferedWriter writer, String line) throws IOException {
        writer.write(line);
        writer.write("\n");
        writer.flush();
    }

    private static String extractTableName(String sql) {
        sql = sql.toUpperCase();
        int idx = sql.indexOf("FROM FILEC.");
        if (idx == -1) return "UNKNOWN";
        int start = idx + 10;
        int end = sql.indexOf(" ", start);
        if (end == -1) end = sql.length();
        return sql.substring(start, end).replaceAll("[^A-Z0-9_]", "");
    }
}