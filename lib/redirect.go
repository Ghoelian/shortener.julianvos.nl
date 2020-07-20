package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// IPInfo describes an IP address
type IPInfo struct {
	City    string
	Country string
	Region  string
}

func Redirect(res http.ResponseWriter, req *http.Request) {
	var err error
	var dbHost = os.Getenv("SQL_HOST")
	var dbUser = os.Getenv("SQL_USER")
	var dbPass = os.Getenv("SQL_PASS")
	var smtpHost = os.Getenv("SMTP_HOST")
	var smtpUser = os.Getenv("SMTP_USER")
	var smtpPass = os.Getenv("SMTP_PASS")
	var ipInfoToken = os.Getenv("IPINFO_TOKEN")
	var ipInfo IPInfo

	var ip = filterIP(req.RemoteAddr)
	var now = time.Now()
	var destination = req.URL.Query()["destination"][0]
	var origin = req.URL.Query()["origin"][0]

	http.Redirect(res, req, destination, 302)
	defer req.Body.Close()

	response, err := http.Get("https://ipinfo.io/" + ip + "/json?token=" + ipInfoToken)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(contents, &ipInfo); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v)/", dbUser, dbPass, dbHost))
	if err != nil {
		log.Fatal(err)
	}

	db.Query(fmt.Sprintf("INSERT INTO shortener.tracking (date, destination, origin, ip, city, country, region) VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v')", now, destination, origin, ip, ipInfo.City, ipInfo.Country, ipInfo.Region))

	defer db.Close()

	sendMail(smtpUser, smtpPass, smtpHost, origin, ip, destination, now)
}

func sendMail(smtpUser string, smtpPass string, smtpHost string, origin string, ip string, destination string, now time.Time) {
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{smtpUser}
	msg := []byte(fmt.Sprintf("To: %v\r\n"+
		"Subject: New click from %v\r\n"+
		"\r\n"+
		"Someone with IP %v clicked on a link to %v from %v at %v\r\n",
		to, origin, ip, destination, origin, now))

	err := smtp.SendMail(smtpHost+":587", auth, smtpUser, to, msg)
	if err != nil {
		log.Fatal(err)
	}
}

func filterIP(ip string) string {
	var filteredIP string

	if string([]rune(ip)[0]) == "[" {
		temp := strings.Split(ip, "[")[1]
		filteredIP = strings.Split(temp, "]")[0]
	} else {
		filteredIP = strings.Split(ip, ":")[0]
	}

	return filteredIP
}
