import java.sql.*;
import java.util.*;

public class JDBCRunner {
    public static void main(String[] args) throws Exception {
        if (args.length < 4) {
            printStatus("error", "Usage: java -cp jt400.jar:. JDBCRunner <host> <user> <pass> <sql|test>");
            return;
        }

        String host = args[0];
        String user = args[1];
        String pass = args[2];

        boolean testConnection = args.length >= 4 && "test".equalsIgnoreCase(args[3]);
        String url = "jdbc:as400://" + host + ";naming=system";

        if (testConnection) {
            try (Connection conn = DriverManager.getConnection(url, user, pass)) {
                printStatus("ok", "Connection successful");
            } catch (Exception ex) {
                printStatus("error", "Connection failed: " + ex.getMessage());
                System.exit(1);
            }
            return;
        }

        String sql = args[3];

        try (Connection conn = DriverManager.getConnection(url, user, pass);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {

            ResultSetMetaData md = rs.getMetaData();
            int colCount = md.getColumnCount();

            // For each row, print NDJSON (one JSON object per line)
            while (rs.next()) {
                Map<String, String> row = new LinkedHashMap<>();
                row.put("SRC_TABLE", extractTableName(sql));
                for (int i = 1; i <= colCount; i++) {
                    String val = rs.getString(i);
                    row.put(md.getColumnName(i), val != null ? val : "");
                }
                System.out.println(toJson(row));
            }
            printStatus("ok", "Query completed");
        } catch (Exception ex) {
            printStatus("error", "Query failed: " + ex.getMessage());
            System.exit(1);
        }
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

    private static void printStatus(String status, String message) {
        System.out.println("{\"status\":\"" + escapeJson(status) + "\",\"message\":\"" + escapeJson(message) + "\"}");
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