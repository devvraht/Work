package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/websocket"
)

const folderPath = `C:\Users\ADMIN\Downloads\drone1`

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // Allow any origin
}

var clients = make(map[*websocket.Conn]bool)

func listFiles() []string {
    files, err := os.ReadDir(folderPath)
    if err != nil {
        log.Println("❌ Error reading folder:", err)
        return []string{}
    }

    var fileList []string
    for _, f := range files {
        if !f.IsDir() {
            fileList = append(fileList, f.Name())
        }
    }
    return fileList
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()
    clients[conn] = true

    log.Println("🔗 WebSocket client connected:", r.RemoteAddr)

    // Send initial file list
    sendFileList(conn)

    // Keep connection alive (no read needed for this simple demo)
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            log.Println("❌ WebSocket closed:", err)
            delete(clients, conn)
            break
        }
    }
}

func sendFileList(conn *websocket.Conn) {
    files := listFiles()
    jsonData, _ := json.Marshal(files)
    conn.WriteMessage(websocket.TextMessage, jsonData)
}

func broadcastFileList() {
    files := listFiles()
    jsonData, _ := json.Marshal(files)

    for conn := range clients {
        err := conn.WriteMessage(websocket.TextMessage, jsonData)
        if err != nil {
            log.Println("❌ Error sending to client:", err)
            conn.Close()
            delete(clients, conn)
        }
    }
}

func watchFolder() {
    previous := make(map[string]bool)

    for {
        current := make(map[string]bool)
        files := listFiles()
        for _, f := range files {
            current[f] = true
        }

        // Detect changes
        if len(current) != len(previous) {
            log.Println("📂 Folder changed, broadcasting update...")
            broadcastFileList()
        }

        previous = current
        time.Sleep(5 * time.Second) // check every 5 seconds
    }
}

func main() {
    // Serve static files for download
    http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(folderPath))))

    // WebSocket endpoint
    http.HandleFunc("/ws", handleWebSocket)

    // Serve the HTML
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })

    // Start folder watcher
    go watchFolder()

    log.Println("🚀 WebSocket server running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
