package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/websocket"
)

var (
	port       = flag.Int("port", 8080, "listen port")
	config     = flag.String("config", "", "config file")
	configDark = flag.String("config-dark", "", "config file for dark theme")
)

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	fmt.Println("port: " + strconv.Itoa(*port))
	fmt.Println("config: " + *config)
	fmt.Println("config-dark: " + *configDark)

	go startServer(*port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Printf("Signal (%s) received, stopping\n", s)
}

func startServer(port int) {
	http.HandleFunc("/", indexHandler)

	http.HandleFunc("/summary.png", summaryHandler)
	http.HandleFunc("/vsummary.png", vsummaryHandler)
	http.HandleFunc("/hsummary.png", hsummaryHandler)
	http.HandleFunc("/top.png", topHandler)
	http.HandleFunc("/years.png", yearHandler)
	http.HandleFunc("/months.png", monthHandler)
	http.HandleFunc("/days.png", dayHandler)
	http.HandleFunc("/hours.png", hourHandler)
	http.HandleFunc("/hoursgraph.png", hourgraphHandler)
	http.HandleFunc("/five.png", fiveHandler)
	http.HandleFunc("/fivegraph.png", fivegraphHandler)

	http.HandleFunc("/summary", summaryPageHandler)
	http.HandleFunc("/top", topPageHandler)
	http.HandleFunc("/years", yearPageHandler)
	http.HandleFunc("/months", monthPageHandler)
	http.HandleFunc("/days", dayPageHandler)
	http.HandleFunc("/hours", hourPageHandler)
	http.HandleFunc("/five", fivePageHandler)

	http.Handle("/live", websocket.Handler(liveHandler))
	http.ListenAndServe("[::]:"+strconv.Itoa(port), nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/summary", http.StatusMovedPermanently)
}

func pageHandler(w http.ResponseWriter, r *http.Request, images ...string) {
	iface := r.URL.Query().Get("iface")
	query := "?"
	if len(iface) > 0 {
		query += "iface=" + iface
	}
	ifList := getIfList()

	// check iface safe
	if !checkIface(iface, ifList) {
		fmt.Fprintf(w, "Interface does not exist: "+iface)
		return
	}

	printPageHeader(w)
	printNav(w, r, ifList)
	fmt.Fprintf(w, "<div class=\"content\">")
	fmt.Fprintf(w, "<div id=\"live\"></div>")
	for _, image := range images {
		fmt.Fprintf(w, "<picture>")
		fmt.Fprintf(w, "<source srcset=\""+image+".png"+query+"&dark=1\" media=\"(prefers-color-scheme: dark)\">")
		fmt.Fprintf(w, "<source srcset=\""+image+".png"+query+"&dark=0\">")
		fmt.Fprintf(w, "<img class=\"light\" src=\""+image+".png"+query+"\" alt=\""+image+".png\">")
		fmt.Fprintf(w, "</picture>")
	}
	fmt.Fprintf(w, "<div>")
	printPageFooter(w)
}

func summaryPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "vsummary")
}

func topPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "top")
}

func yearPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "years")
}

func monthPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "months")
}

func dayPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "days")
}

func hourPageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "hours", "hoursgraph")
}

func fivePageHandler(w http.ResponseWriter, r *http.Request) {
	pageHandler(w, r, "five", "fivegraph")
}

func printPageHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Expires", "0")
	w.Header().Add("Refresh", "300")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head>")
	printCss(w)
	fmt.Fprintf(w, "</head><body>")
}

func printPageFooter(w http.ResponseWriter) {
	printScript(w)
	fmt.Fprintf(w, "</body></html>")
}

func printNav(w http.ResponseWriter, r *http.Request, ifList []string) {
	fmt.Fprintf(w, "<ul class=\"nav\">")

	// interface select
	selectedIface := r.URL.Query().Get("iface")
	fmt.Fprintf(w, "<select onchange=\"location.replace(location.pathname+'?iface='+encodeURIComponent(this.value))\" style=\"width: 120px\">")
	fmt.Fprintf(w, "<option value=\"\">default interface</option>")
	for _, iface := range ifList {
		if len(iface) > 0 {
			iface = strings.TrimSpace(iface)
			if iface == selectedIface {
				fmt.Fprintf(w, "<option selected>"+iface+"</option>")
			} else {
				fmt.Fprintf(w, "<option>"+iface+"</option>")
			}
		}
	}
	fmt.Fprintf(w, "</select>")

	query := ""
	if len(selectedIface) > 0 {
		query = "?iface=" + selectedIface
	}
	fmt.Fprint(w, "<li><a href=\"summary"+query+"\">Summary</a></li>")
	fmt.Fprint(w, "<li><a href=\"top"+query+"\">Top</a></li>")
	fmt.Fprint(w, "<li><a href=\"years"+query+"\">Years</a></li>")
	fmt.Fprint(w, "<li><a href=\"months"+query+"\">Months</a></li>")
	fmt.Fprint(w, "<li><a href=\"days"+query+"\">Days</a></li>")
	fmt.Fprint(w, "<li><a href=\"hours"+query+"\">Hours</a></li>")
	fmt.Fprint(w, "<li><a href=\"five"+query+"\">Five Minutes</a></li>")
	fmt.Fprintf(w, "</ul>")
}

func printCss(w http.ResponseWriter) {
	fmt.Fprint(w, `<style>
body, input, select {
  color: #222;
  background: #fff;
  font: 100% system-ui;
}

a {
  color: #0033cc;
}

@media (prefers-color-scheme: dark) {
  body, input, select {
    color: #eee;
    background: #121212;
  }

  a {
    color: #809fff;
  }
}

.nav {
	display: inline-block;
	margin-right: 42px;
}

.content {
	display: inline-block;
	vertical-align: top;
}

.content img {
	margin-top: 16px;
	display: block;
}

#live {
    font-family: Courier, monospace;
}
</style>`)
}

func printScript(w http.ResponseWriter) {
	fmt.Fprint(w, `<script>
(function connect() {
	const socket = new WebSocket("ws://"+location.host+"/live"+location.search);
	socket.addEventListener("open", function (event) {
		socket.send("Hello!");
	});
	socket.addEventListener("message", function (event) {
		document.getElementById("live").innerText = event.data;
	});
	socket.addEventListener("close", function (event) {
		console.log('Socket is closed. Reconnect will be attempted in 10 second.', event.reason);
		document.getElementById("live").innerText = "";
		delete socket;
		setTimeout(function() {
			connect();
		}, 10000);
	});
})();
</script>`)
}

func imageHandler(w http.ResponseWriter, r *http.Request, args ...string) {
	iface := r.URL.Query().Get("iface")
	// check iface safe
	if !checkIface(iface, nil) {
		fmt.Fprintf(w, "Interface does not exist: "+iface)
		return
	}

	dark := r.URL.Query().Get("dark")
	if len(iface) > 0 {
		args = append(args, "-i", iface)
	}
	if dark == "1" && len(*configDark) > 0 {
		args = append(args, "--config", *configDark)
	} else if len(*config) > 0 {
		args = append(args, "--config", *config)
	}
	args = append(args, "-o", "-")
	cmd := exec.Command("vnstati", args...)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Write(stdout)
}

func summaryHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-s")
}

func vsummaryHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-vs")
}

func hsummaryHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-hs")
}

func topHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-t")
}

func yearHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-y")
}

func monthHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-m")
}

func dayHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-d")
}

func hourHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-h")
}

func hourgraphHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-hg")
}

func fiveHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-5")
}

func fivegraphHandler(w http.ResponseWriter, r *http.Request) {
	imageHandler(w, r, "-5g")
}

func getIfList() []string {
	cmd := exec.Command("vnstat", "--iflist", "1")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	iflist := strings.Split(string(stdout), "\n")

	var result []string
	for _, str := range iflist {
		if len(str) > 0 {
			result = append(result, str)
		}
	}
	return result
}

func checkIface(iface string, ifList []string) bool {
	if len(iface) == 0 {
		return true
	}
	if ifList == nil {
		ifList = getIfList()
	}
	for _, a := range ifList {
		if a == iface {
			return true
		}
	}
	return false
}

func liveHandler(ws *websocket.Conn) {
	defer ws.Close()

	var err error
	var reply string
	if err = websocket.Message.Receive(ws, &reply); err != nil {
		fmt.Println("receive failed:", err)
		return
	}
	fmt.Println("reveived from client: " + reply)

	iface := ws.Request().URL.Query().Get("iface")
	if !checkIface(iface, nil) {
		fmt.Println("Interface does not exist: ", iface)
		return
	}

	args := []string{"-l"}
	if len(iface) > 0 {
		args = append(args, "-i", iface)
	}
	if len(*config) > 0 {
		args = append(args, "--config", *config)
	}
	cmd := exec.Command("vnstat", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = cmd.Start()
	defer cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buf := make([]byte, 1024)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		read, err := stdout.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		str := string(buf[:read])
		i := strings.Index(str, "rx:")
		if i != -1 {
			if err = websocket.Message.Send(ws, str[i:]); err != nil {
				fmt.Println("send failed:", err)
				return
			}
		}
	}
}
