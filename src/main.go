package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

const (
    versionPath  = "/version"
    healthzPath  = "/healthz"
    defaultPath  = "/"
    defaultPort  = ":9000"
    certFile     = "/tmp/tls/tls.crt"
    privateKey  = "/tmp/tls/tls.key"
    maintenanceMessage = "The service is currently under regular maintenance, please try again later"
    maintenanceCode = "WARN_UNAVAILABLE_MAINTENANCE"
)

var (
    upgrader = websocket.Upgrader{} // use default options
    maintenanceResponse = jsonResponse(RespM{
        Code: maintenanceCode,
        Message: "Maintenance",
        Details: maintenanceMessage,
    })
)

//RespM message
type RespM struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details"`
}

func jsonResponse(resp interface{}) []byte {
    js, err := json.Marshal(resp)
    if err != nil {
        log.Println("json:", err)
        return []byte{}
    }
    return js
}

func version(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("gohttps v1.0.0"))
}

func healthz(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
    wss := r.Header.Get("Upgrade")
    if wss != "" {
        upgrader.CheckOrigin = func(r *http.Request) bool { return true }
        c, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Print("upgrade:", err)
            return
        }
        defer c.Close()
        for {
            mt, message, err := c.ReadMessage()
            if err != nil {
                log.Println("read:", err)
                break
            }
            log.Printf("recv: %s %s", message)
            err = c.WriteMessage(mt, maintenanceResponse)
            if err != nil {
                log.Println("write:", err)
                break
            }
        }
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusServiceUnavailable)
    io.WriteString(w, string(maintenanceResponse))
}

func main() {
    http.HandleFunc(versionPath, version)
    http.HandleFunc(healthzPath, healthz)
    http.HandleFunc(defaultPath, handleDefault)

    log.Fatal(http.ListenAndServeTLS(defaultPort, certFile, privateKey, nil))
}
