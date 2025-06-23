import java.sql.*;
import java.util.*;

public class JDBCRunner {
    public static void main(String[] args) throws Exception {
        if (args.length < 4) {
            System.err.println("Usage: java -cp jt400.jar:. JDBCRunner <host> <user> <pass> <sql>");
            return;
        }

        String host = args[0];
        String user = args[1];
        String pass = args[2];
        String sql = args[3];

        String srcTable = extractTableName(sql);
        String url = "jdbc:as400://" + host + ";naming=system";

        try (Connection conn = DriverManager.getConnection(url, user, pass);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {

            ResultSetMetaData md = rs.getMetaData();
            int colCount = md.getColumnCount();

            // For each row, print NDJSON (one JSON object per line)
            while (rs.next()) {
                Map<String, String> row = new LinkedHashMap<>();
                row.put("SRC_TABLE", srcTable);
                for (int i = 1; i <= colCount; i++) {
                    String val = rs.getString(i);
                    row.put(md.getColumnName(i), val != null ? val : "");
                }
                System.out.println(toJson(row));
            }
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