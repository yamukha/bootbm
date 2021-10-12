package main

import (
    log "github.com/sirupsen/logrus"
    "github.com/tarm/serial"
    toml "github.com/BurntSushi/toml"
    "fmt"
    "time"
    "os"
    "strings"
)

type config struct {
    Title string    `toml:"title"`
    T0     marker   `toml:"t0"`
    T1     marker   `toml:"t1"`
    T2     marker   `toml:"t2"`
    Iout   iface    `toml:"iout"`
    Login  login    `toml:"login"`
    Log    level `toml:"loglevel"`
}

type login struct {
  Login string    `toml:"login"`
  Pswd  string    `toml:"pswd"`
  Password string `toml:"password"`
  User string     `toml:"user"`
}

type marker struct {
    Marker    string `toml:"marker"`
}

type iface struct {
    Log       string `toml:"file"`
    Port      string `toml:"port"`
    Baudrate  int    `toml:"br"`
    OS        string `toml:"os"`
    Delay     int    `toml:"delay"`
    Timeout   int    `toml:"timeout"`
}

type level struct {
    Level    string `toml:"level"`
}

func main() {

    if 2 > len(os.Args) {
       fmt.Println ("usage:")
       fmt.Println ("bootbm  conf.toml") 
       os.Exit(2)
    }

    var conf config
    if _, err := toml.DecodeFile(os.Args[1], &conf); err != nil {
        fmt.Println(err)
     return
    }

    logFile := conf.Iout.Log
    to := conf.Iout.Timeout
    portname := conf.Iout.Port
    baudrate := conf.Iout.Baudrate
    prompt   := byte(0xd)
    prompt2  := byte(0xa)

    fmt.Println ("try to open ", portname, baudrate, "8N1 timeout", to, "ms")
    c := &serial.Config{Name: portname, Baud: baudrate, ReadTimeout: time.Millisecond * time.Duration(to),}
    port, err := serial.OpenPort(c)
    if err != nil {
        log.Fatal(err, ": ", portname)
    }
    defer port.Close()
    port.Flush()

    f, err := os.Create(logFile)
    if err != nil {
        log.Fatal(err, ": ", logFile)
    }
    defer f.Close()

    fmt.Println ("go ...")
    var rx_buf []byte  // rx_buf := make([]byte, 0)
    buf := make([]byte, 128)
    var t0, t1, t2 int64

    for {

        n, err := port.Read(buf)

        if err != nil {
            log.Fatal(err)
        }

        if n > 0 {
            slc := make([]byte, n)
            copy(slc, buf[:n]) 
            rx_buf = append (rx_buf, slc...);
 
            if  len(rx_buf) >= len(conf.T0.Marker) && strings.Contains(string(rx_buf), conf.T0.Marker) {
                t0 =  time.Now().UnixNano()/1000
                s := fmt.Sprintf("t0: %d us\n", 0)
                _, err := f.Write([]byte(s))
                if err != nil {
                    log.Fatal(err, ": ", logFile)
                }
                rx_buf = nil
            }

            if  len(rx_buf) >= len(conf.T1.Marker) && strings.Contains(string(rx_buf), conf.T1.Marker) {
                t1 =  time.Now().UnixNano()/1000
                s := fmt.Sprintf("t1: %d us\n", t1 - t0)
                _, err := f.Write([]byte(s))
                if err != nil {
                    log.Fatal(err, ": ", logFile)
                }
                v := []byte(conf.Login.User)
                port.Write(v)
                rx_buf = nil
            }

            if len(rx_buf) >= len(conf.T2.Marker) && strings.Contains(string(rx_buf), conf.T2.Marker)  {
                t2 =  time.Now().UnixNano()/1000
                s := fmt.Sprintf("t2: %d us\n",t2 - t0)
                _, err := f.Write([]byte(s))
                if err != nil {
                    log.Fatal(err, ": ", logFile)
                }
                v := []byte(conf.Login.Password)
                port.Write(v)
                rx_buf = nil
            }

            if len(rx_buf) > 1 && (rx_buf[len(rx_buf)-1] == prompt || rx_buf[len(rx_buf)-1] == prompt2 ) {
                fmt.Print(strings.TrimLeft(string(rx_buf), " "))
                rx_buf = nil
            }
        }
    }
}