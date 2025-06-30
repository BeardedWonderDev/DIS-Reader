import java.io.*;
import java.net.*;
import java.sql.*;
import java.util.*;

public class JDBCRunner {
    private static Connection conn = null;
    private static String host;
    private static String user;
    private static String pass;
    private static String url;

    public static void main(String[] args) throws Exception {
        if (args.length < 4) {
            System.err.println("Usage: java -cp jt400.jar:. JDBCRunner <host> <user> <pass> <port>");
            System.exit(1);
        }

        host = args[0];
        user = args[1];
        pass = args[2];
        url = "jdbc:as400://" + host + ";naming=system";
        int port = Integer.parseInt(args[3]);

        ServerSocket serverSocket = new ServerSocket(port);
        System.out.println("{\"status\":\"ok\",\"message\":\"Server started on port " + port + "\"}");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                if (conn != null && !conn.isClosed()) {
                    conn.close();
                    conn = null;
                }
                serverSocket.close();
                System.out.println("{\"status\":\"ok\",\"message\":\"Server stopped\"}");
            } catch (IOException | SQLException e) {
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
        try (
            BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), "UTF-8"));
            BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), "UTF-8"))
        ) {
            String line;
            while ((line = reader.readLine()) != null) {
                Map<String, String> cmdMap = parseJson(line);
                if (cmdMap == null || !cmdMap.containsKey("cmd")) {
                    writeLine(writer, "{\"status\":\"error\",\"message\":\"Invalid command format\"}");
                    break;
                }

                String cmd = cmdMap.get("cmd").toLowerCase();
                switch (cmd) {
                    case "connect":
                        handleConnect(writer);
                        break;
                    case "disconnect":
                        handleDisconnect(writer);
                        socket.close();
                        return;
                    case "query":
                        if (!cmdMap.containsKey("sql")) {
                            writeLine(writer, "{\"status\":\"error\",\"message\":\"Missing 'sql' field in query command\"}");
                        } else {
                            handleQuery(writer, cmdMap.get("sql"), socket);
                        }
                        break;
                    default:
                        writeLine(writer, "{\"status\":\"error\",\"message\":\"Unknown command: " + escapeJson(cmd) + "\"}");
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

    private static void handleConnect(BufferedWriter writer) throws IOException {
        try {
            if (conn != null && !conn.isClosed()) {
                writeLine(writer, "{\"status\":\"ok\",\"message\":\"Already connected\"}");
                return;
            }
            conn = DriverManager.getConnection(url, user, pass);
            writeLine(writer, "{\"status\":\"ok\",\"message\":\"Connection successful\"}");
        } catch (Exception ex) {
            writeLine(writer, "{\"status\":\"error\",\"message\":\"Connection failed: " + escapeJson(ex.getMessage()) + "\"}");
            conn = null;
        }
    }

    private static void handleDisconnect(BufferedWriter writer) throws IOException {
        try {
            if (conn != null && !conn.isClosed()) {
                conn.close();
                conn = null;
                writeLine(writer, "{\"status\":\"ok\",\"message\":\"Disconnected\"}");
            } else {
                writeLine(writer, "{\"status\":\"ok\",\"message\":\"No active connection\"}");
            }
        } catch (Exception ex) {
            writeLine(writer, "{\"status\":\"error\",\"message\":\"Error during disconnect: " + escapeJson(ex.getMessage()) + "\"}");
        }
    }

    private static void handleQuery(BufferedWriter writer, String sql, Socket socket) throws IOException {
        if (conn == null) {
            writeLine(writer, "{\"status\":\"error\",\"message\":\"Not connected\"}");
            writer.flush();
            return;
        }
        try (Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {

            ResultSetMetaData md = rs.getMetaData();
            int colCount = md.getColumnCount();

            while (rs.next()) {
                Map<String, String> row = new LinkedHashMap<>();
                row.put("SRC_TABLE", extractTableName(sql));
                for (int i = 1; i <= colCount; i++) {
                    String val = rs.getString(i);
                    row.put(md.getColumnName(i), val != null ? val : "");
                }
                writeLine(writer, toJson(row));
            }
            writeLine(writer, "{\"status\":\"ok\",\"message\":\"Query completed\"}");
            writeLine(writer, "{\"status\":\"done\"}");
            writer.flush();
            return;
        } catch (Exception ex) {
            writeLine(writer, "{\"status\":\"error\",\"message\":\"Query failed: " + escapeJson(ex.getMessage()) + "\"}");
            writer.flush();
            try {
                if (conn != null && conn.isClosed()) {
                    conn = null; // connection lost, reset
                }
            } catch (Exception ignore) {}
            return;
        }
    }

    private static void writeLine(BufferedWriter writer, String line) throws IOException {
        writer.write(line);
        writer.write("\n");
        writer.flush();
    }

    // Simple manual JSON serializer (handles strings only, escapes quotes/backslashes)
    private static String toJson(Map<String, String> map) {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        boolean first = true;
        for (Map.Entry<String, String> entry : map.entrySet()) {
            if (!first) sb.append(",");
            sb.append("\"").append(escapeJson(entry.getKey())).append("\":");
            sb.append("\"").append(escapeJson(entry.getValue())).append("\"");
            first = false;
        }
        sb.append("}");
        return sb.toString();
    }

    private static String escapeJson(String s) {
        if (s == null) return "";
        return s.replace("\\", "\\\\")
                .replace("\"", "\\\"")
                .replace("\b", "\\b")
                .replace("\f", "\\f")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }

    private static Map<String, String> parseJson(String json) {
        // Very simple JSON parser expecting flat JSON objects with string values only
        // Format: {"key":"value",...}
        Map<String, String> map = new HashMap<>();
        json = json.trim();
        if (!json.startsWith("{") || !json.endsWith("}")) return null;
        json = json.substring(1, json.length() - 1).trim();
        if (json.isEmpty()) return map;

        int i = 0;
        while (i < json.length()) {
            // parse key
            if (json.charAt(i) != '"') return null;
            int keyStart = i + 1;
            int keyEnd = json.indexOf('"', keyStart);
            if (keyEnd == -1) return null;
            String key = unescapeJson(json.substring(keyStart, keyEnd));
            i = keyEnd + 1;

            // skip colon
            while (i < json.length() && Character.isWhitespace(json.charAt(i))) i++;
            if (i >= json.length() || json.charAt(i) != ':') return null;
            i++;
            while (i < json.length() && Character.isWhitespace(json.charAt(i))) i++;
            if (i >= json.length()) return null;

            // parse value
            if (json.charAt(i) != '"') return null;
            int valStart = i + 1;
            int valEnd = valStart;
            boolean escape = false;
            StringBuilder valBuilder = new StringBuilder();
            while (valEnd < json.length()) {
                char c = json.charAt(valEnd);
                if (escape) {
                    valBuilder.append(c);
                    escape = false;
                } else {
                    if (c == '\\') {
                        escape = true;
                    } else if (c == '"') {
                        break;
                    } else {
                        valBuilder.append(c);
                    }
                }
                valEnd++;
            }
            if (valEnd >= json.length()) return null;
            String value = unescapeJson(valBuilder.toString());
            i = valEnd + 1;

            map.put(key, value);

            // skip comma or end
            while (i < json.length() && Character.isWhitespace(json.charAt(i))) i++;
            if (i == json.length()) break;
            if (json.charAt(i) == ',') {
                i++;
                while (i < json.length() && Character.isWhitespace(json.charAt(i))) i++;
            } else {
                return null;
            }
        }
        return map;
    }

    private static String unescapeJson(String s) {
        StringBuilder sb = new StringBuilder();
        boolean escape = false;
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            if (escape) {
                switch (c) {
                    case 'b': sb.append('\b'); break;
                    case 'f': sb.append('\f'); break;
                    case 'n': sb.append('\n'); break;
                    case 'r': sb.append('\r'); break;
                    case 't': sb.append('\t'); break;
                    case '\\': sb.append('\\'); break;
                    case '"': sb.append('"'); break;
                    default: sb.append(c); break;
                }
                escape = false;
            } else if (c == '\\') {
                escape = true;
            } else {
                sb.append(c);
            }
        }
        return sb.toString();
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