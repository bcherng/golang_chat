package lib.chat;

import java.io.*;
import java.net.*;

class ChatClient {
    public static void main(String... args) throws IOException {
        String host = "localhost";
        int port = 6666;

        if (args.length == 2) {
            host = args[0];
            port = Integer.parseInt(args[1]);
        }

        try (
            var s = new Socket(host, port);
            var in = new BufferedReader(new InputStreamReader(s.getInputStream()));
            var out = new PrintWriter(new OutputStreamWriter(s.getOutputStream()), true);
            var stdin = new BufferedReader(new InputStreamReader(System.in));
        ) {
            // Create a separate thread for reading input from the server (in)
            Thread clientThread = new Thread(() -> {
                String line;
                try {
                    System.out.print(">");
                    while ((line = stdin.readLine()) != null) {
                        out.println(line);
                    }
                } catch (IOException e) {
                    e.printStackTrace();
                }
            });
            clientThread.start();

            String reply;
            while (true) {
                if ((reply = in.readLine()) == null) {
                    break;
                }
                System.out.println(reply);
                System.out.print(">");
            }
        }
    }
}
