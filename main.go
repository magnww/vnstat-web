package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	port = flag.Int("port", 8080, "listen port")
)

func main() {
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
	http.ListenAndServe("[::]:"+strconv.Itoa(port), nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/summary", http.StatusMovedPermanently)
}

func pageHandler(w http.ResponseWriter, r *http.Request, images ...string) {
	printPageHeader(w)
	printNav(w)
	fmt.Fprintf(w, "<div class=\"content\">")
	for _, image := range images {
		fmt.Fprintf(w, "<img src=\""+image+".png\" alt=\""+image+".png\">")
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
	fmt.Fprintf(w, "</body></html>")
}

func printNav(w http.ResponseWriter) {
	fmt.Fprintf(w, "<ul class=\"nav\">")
	fmt.Fprint(w, "<li><a href=\"summary\">Summary</a></li>")
	fmt.Fprint(w, "<li><a href=\"top\">Top</a></li>")
	fmt.Fprint(w, "<li><a href=\"years\">Years</a></li>")
	fmt.Fprint(w, "<li><a href=\"months\">Months</a></li>")
	fmt.Fprint(w, "<li><a href=\"days\">Days</a></li>")
	fmt.Fprint(w, "<li><a href=\"hours\">Hours</a></li>")
	fmt.Fprint(w, "<li><a href=\"five\">Five Minutes</a></li>")
	fmt.Fprintf(w, "</ul>")
}

func printCss(w http.ResponseWriter) {
	fmt.Fprint(w, `<style>
body {
  color: #222;
  background: #fff;
  font: 100% system-ui;
}
a {
  color: #0033cc;
}

@media (prefers-color-scheme: dark) {
  body {
    color: #eee;
    background: #121212;
  }

  body a {
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
	display: block;
	margin-top: 16px;
}
</style>`)
}

func imageHandler(w http.ResponseWriter, r *http.Request, args ...string) {
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
