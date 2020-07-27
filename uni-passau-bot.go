package unipassaubot

import (
	"bufio"
	"encoding/csv"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/keybase/go-logging"
	"golang.org/x/text/encoding/charmap"
	tb "gopkg.in/tucnak/telebot.v2"
)

var mensaBotLog = logging.MustGetLogger("mensaBot")

// Global Variables
// Matrix Slice for food handling (should be replaced in future??)
var values [][]string

//var nextvalues [][]string

// uniPassauBot handles all the legacy uni-passau-bot code for telegram
func uniPassauBot() {

	botquit := make(chan bool) // channel for quitting of bot

	// catch os signals like sigterm and interrupt
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			mensaBotLog.Info("Interruption Signal received, shutting down...")
			exit(botquit)
		case syscall.SIGTERM:
			botquit <- true
		}
	}()

	// check for and read config variable, then create bot object
	token := os.Getenv("UNIPASSAUBOT_TOKEN")
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		mensaBotLog.Error("Error starting Uni Passau Bot", err)
		return
	}

	// init reply keyboard
	replyBtn := tb.ReplyButton{Text: "Food for today"}
	replyBtn2 := tb.ReplyButton{Text: "Food for tomorrow"}
	replyBtn3 := tb.ReplyButton{Text: "Food for the week"}
	replyKeys := [][]tb.ReplyButton{
		{replyBtn, replyBtn2}, {replyBtn3}}

	// Command Handlers
	// handle special keyboard commands
	b.Handle(&replyBtn, func(m *tb.Message) {
		if get("isCorona") != "true" {
			_, _ = b.Send(m.Sender, foodtoday(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
	})
	b.Handle(&replyBtn2, func(m *tb.Message) {
		if get("isCorona") != "true" {
			_, _ = b.Send(m.Sender, foodtomorrow(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
	})
	b.Handle(&replyBtn3, func(m *tb.Message) {
		if get("isCorona") != "true" {
			_, _ = b.Send(m.Sender, foodweek(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
	})
	// handle standard text commands
	b.Handle("/hello", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Hi! How are you?", tb.ModeMarkdown)
		printInfo(m)
	})
	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Hallo! Ich bin der inoffizielle ChatBot der Uni Passau! Was kann ich dir Gutes tun?\nWenn du Hilfe benötigst benutze einfach /help!\nSolltest du den Mensa- und Stundenplan in einer App wollen, schreibe /app für mehr Informationen", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
		printInfo(m)
	})
	b.Handle("/app", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Du kannst dir die Android-App im [Play Store](https://play.google.com/store/apps/details?id=studip_uni_passau.femtopedia.de.unipassaustudip) gratis herunterladen.\nHinweis: Diese App wird von einer anderen Person entwickelt, bitte kontaktiere den App-Entwickler für Support!", tb.ModeMarkdown)
		printInfo(m)

	})
	b.Handle("/help", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "Information about the Bot is in the Description\nAvailable Commands are:\n*/help* - Show this help\n*/food* - Get Information for the food TODAY in the Uni Passau\n*/foodtomorrow* - Get Information for the food TOMORROW in the Uni Passau\n*/foodweek* - Get Information for the wood this WEEK in the Uni Passau\n*/contact* - Contact the bot maintainer for requests and bug reports\n*/app* - More Information for an useful Android-App for studip", tb.ModeMarkdown)
		printInfo(m)
	})
	b.Handle("/food", func(m *tb.Message) {
		if get("isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, foodtoday())
				mensaBotLog.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, foodtoday(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
		//printAnswer(foodtoday())
	})
	b.Handle("/foodtomorrow", func(m *tb.Message) {
		if get("isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, foodtomorrow())
				mensaBotLog.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, foodtomorrow(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
	})
	b.Handle("/foodweek", func(m *tb.Message) {
		if get("isCorona") != "true" {
			if !m.Private() {
				_, _ = b.Send(m.Chat, foodweek())
				//_, _ = b.Send(m.Chat, "This command is temporarily disabled.")
				mensaBotLog.Info("Group Message:")
			} else {
				_, _ = b.Send(m.Sender, foodweek(), &tb.ReplyMarkup{ReplyKeyboard: replyKeys}, tb.ModeMarkdown)
				//_, _ = b.Send(m.Sender, "This command is temporarily disabled.")
			}
		} else {
			_, _ = b.Send(m.Chat, "Sorry, it's Corona time! 😔")
		}
		printInfo(m)
	})
	b.Handle("/contact", func(m *tb.Message) {
		sendstring := ""
		if m.Text == "/contact" {
			_, _ = b.Send(m.Sender, "For requests and bug reports just add your message to the _/contact_ command.", tb.ModeMarkdown)
		} else {
			_, _ = b.Send(m.Sender, "Sending Message to the Bot Maintainer...")
			tionis := tb.Chat{ID: 248533143}
			sendstring = "Message by " + m.Sender.FirstName + " " + m.Sender.LastName + "\nID: " + strconv.Itoa(m.Sender.ID) + " Username: " + m.Sender.Username + "\n- - - - -\n" + strings.TrimPrefix(m.Text, "/contact ")
			_, _ = b.Send(&tionis, sendstring)
		}
		printInfo(m)
		printAnswer(sendstring)
	})
	b.Handle("/send", func(m *tb.Message) {
		if m.Sender.ID == 248533143 {
			s1 := strings.TrimPrefix(m.Text, "/send ")
			s := strings.Split(s1, "$")
			recID, _ := strconv.ParseInt(s[0], 10, 64)
			rec := tb.Chat{ID: recID}
			_, _ = b.Send(&rec, s[1])
		} else {
			_, _ = b.Send(m.Sender, "You are not authorized to execute this command!")
			printInfo(m)
		}
	})
	b.Handle("Danke", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_Gern geschehen!_", tb.ModeMarkdown)
		printInfo(m)
		printAnswer("_Gern geschehen!_")
	})
	b.Handle("Thanks", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_It's a pleasure!_", tb.ModeMarkdown)
		printInfo(m)
		printAnswer("_It's a pleasure!_")
	})
	b.Handle("/ping", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, "_pong_", tb.ModeMarkdown)
		printInfo(m)
		printAnswer("_pong_")
	})
	b.Handle(tb.OnAddedToGroup, func(m *tb.Message) {
		mensaBotLog.Info("Group Message:")
		printInfo(m)
	})
	b.Handle(tb.OnText, func(m *tb.Message) {
		sendstring := "Unknown Command - use help to get a list of available commands"
		if !m.Private() {
			mensaBotLog.Info("Message from Group:")
		} else {
			_, _ = b.Send(m.Sender, sendstring)
		}
		printInfo(m)
		printAnswer(sendstring)
	})

	// Graceful Shutdown (botquit)
	go func() {
		<-botquit
		b.Stop()
		mensaBotLog.Info("Bot was stopped")
		os.Exit(3)
	}()

	// init preparations
	initArray()

	// print startup message
	mensaBotLog.Info("Starting up...")
	b.Start()
}

// Initializes the Array
func initArray() {
	updateFile()
	r := csv.NewReader(bufio.NewReader(openactFile()))
	values = nil
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		// Build Slice by appending every line
		values = append(values, record)
	}
}

/*func initNextArray() {
	// Check if exists
	// checks if new file has to be downloaded and does so - does also remove the old file
	loc, _ := time.LoadLocation("Europe/Berlin")
	_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
	if thisWeek > 51 {
		logger.log("[UniPassauBot] Next week is in next year, this method should not have been executed!")
		return
	}
	nextweekstring := strconv.Itoa(thisWeek + 1)
	if _, err := os.Stat(nextweekstring + ".csv"); os.IsNotExist(err) {

		// No actual file found
		logger.log("[UniPassauBot] " + "No File for next week found - starting download --- ")
		err := downloadFile(nextweekstring)
		if err != nil {
			panic(err)
		}
	} else {
		// logger.log("Up-to-date CSV-File found!")
	}

	r := csv.NewReader(bufio.NewReader(openactFile()))
	nextvalues = nil
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		// Build Slice by appending every line
		nextvalues = append(nextvalues, record)
	}
}*/

// returns a string to send on telegram of the food today
func foodtoday() string {
	// returns the string to print to user who requested the mensa plan
	// reads actual file
	updateFile()

	loc, _ := time.LoadLocation("Europe/Berlin")
	t := time.Now().In(loc)
	ts := t.Format("02.01.2006")
	var daynum int

	for i := 1; i < 8; i++ {
		if weekDate(i) == ts {
			daynum = i
			break
		}
	}

	day := "*Essen am "

	switch daynum {
	case 1:
		day = day + "Montag:* 😋\n"
	case 2:
		day = day + "Dienstag:* 😋\n"
	case 3:
		day = day + "Mittwoch:* 😋\n"
	case 4:
		day = day + "Donnerstag:* 😋\n"
	case 5:
		day = day + "Freitag:* 😋\n"
	case 6, 7:
		day = "_Kein Essen heute!_ 😩"
	default:
		day = "An error occurred, please contact the administrator"
	}

	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(daynum) {
			if len(values[i]) >= 6 {
				day = day + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
			} else {
				day = day + "Error in this line\n"
			}
		}
	}

	return day
}

// returns a string to send on telegram of the food tomorrow
func foodtomorrow() string {
	// returns the string to print to user who requested the mensa plan
	// reads actual file
	updateFile()

	loc, _ := time.LoadLocation("Europe/Berlin")
	t := time.Now().In(loc)
	ts := t.Format("02.01.2006")
	var daynum int

	for i := 1; i < 8; i++ {
		if weekDate(i) == ts {
			daynum = i
			break
		}
	}
	daynum++
	if daynum == 8 {
		// Code not ready yet
		return "This only works weekdays, will be implemented soon!"
		/*
			// Here Code for next week
			loc, _ := time.LoadLocation("Europe/Berlin")
			_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
			//nextweekstring := strconv.Itoa(thisWeek + 1)
			//downloadFile(nextweekstring)
			if thisWeek < 51 {
				initNextArray()
			} else {
				return "Error! - Not implemented yet!"
			}
			// Verarbeitung
			// Has to be monday as it would else trigger a another part of the code
			daynum = 1
			day := "*Essen am Montag:* 😋\n"
			for i := 1; i < len(nextvalues); i++ {
				if nextvalues[i][0] == nextWeekDate(daynum) {
					if len(nextvalues[i]) >= 6 {
						day = day + nextvalues[i][2] + ": " + delInf(nextvalues[i][3]) + " - " + transcor(nextvalues[i][6]) + " €\n"
					} else {
						day = day + "Error in this line\n"
					}
				}
			}

			return day*/
	} else if daynum > 8 {
		return "An Error occurred please contact the administrator"
	}

	day := "*Essen am "

	switch daynum {
	case 1:
		day = day + "Montag:* 😋\n"
	case 2:
		day = day + "Dienstag:* 😋\n"
	case 3:
		day = day + "Mittwoch:* 😋\n"
	case 4:
		day = day + "Donnerstag:* 😋\n"
	case 5:
		day = day + "Freitag:* 😋\n"
	case 6, 7:
		day = "_Kein Essen morgen!_ 😩"
	default:
		day = "An error occurred, please contact the administrator"
	}

	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(daynum) {
			if len(values[i]) >= 6 {
				day = day + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
			} else {
				day = day + "Error in this line\n"
			}
		}
	}

	return day
}

// As transformFile() also changes all the commas in the prices to semicolons this func does the opposite
func transcor(input string) string {
	output := strings.Replace(input, ";", ",", -1)
	return output
}

// returns a string to send on telegram of the food for the week
func foodweek() string {
	// reads actual file
	updateFile()

	var Mo, Di, Mi, Do, Fr string
	Mo = "*Montag*:\n"
	Di = "*Dienstag*:\n"
	Mi = "*Mittwoch*:\n"
	Do = "*Donnerstag*:\n"
	Fr = "*Freitag*:\n"
	dayint := 1
	// Check how long the list for the day is and add it to the string
	for i := 1; i < len(values); i++ {
		if values[i][0] == weekDate(dayint) {
			switch dayint {
			case 1:
				if len(values[i]) >= 6 {
					Mo = Mo + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Mo = Mo + "Error in this line\n"
				}
			case 2:
				if len(values[i]) >= 6 {
					Di = Di + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Di = Di + "Error in this line\n"
				}
			case 3:
				if len(values[i]) >= 6 {
					Mi = Mi + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Mi = Mi + "Error in this line\n"
				}
			case 4:
				if len(values[i]) >= 6 {
					Do = Do + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Do = Do + "Error in this line\n"
				}
			case 5:
				if len(values[i]) >= 6 {
					Fr = Fr + values[i][2] + ": " + delInf(values[i][3]) + " - " + transcor(values[i][6]) + " €\n"
				} else {
					Fr = Fr + "Error in this line\n"
				}
			}
		} else {
			dayint++
		}
	}

	s := []string{Mo, Di, Mi, Do, Fr}
	return strings.Join(s, "\n")
}

// WeekDate returns the date for an specific day
func weekDate(day int) string {
	// Start from the middle of the year:
	loc, _ := time.LoadLocation("Europe/Berlin")
	currentTime := time.Now().In(loc)
	_, week := time.Now().In(loc).UTC().ISOWeek()
	t := time.Date(currentTime.Year(), 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)
	ret := t.AddDate(0, 0, day-1)

	return ret.Format("02.01.2006")
}

/*func nextWeekDate(day int) string {
	// Start from the middle of the year:
	loc, _ := time.LoadLocation("Europe/Berlin")
	currentTime := time.Now().In(loc)
	_, week := time.Now().In(loc).UTC().ISOWeek()
	week++
	t := time.Date(currentTime.Year(), 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)
	ret := t.AddDate(0, 0, day-1)

	return ret.Format("02.01.2006")
}*/

// Delete the symbols in the brackets at the end of the string (in this case the allergic info)
func delInf(input string) string {
	reg := regexp.MustCompile(`\(.*\)`)
	return reg.ReplaceAllString(input, "${1}")
}

// Transforms a given file from iso to utf
func isoToUTF(path string) {
	// Change the encoding and save file under non .tmp name
	f, err := os.Open(path + ".tmp")
	if err != nil {
		mensaBotLog.Error("[UniPassauBot] Error opening file for isoToUTF: ", err)
	}
	out, err := os.Create(path)
	if err != nil {
		mensaBotLog.Error("[UniPassauBot] Error creating path for isoToUTF: ", err)
	}
	r := charmap.ISO8859_1.NewDecoder().Reader(f)
	_, err = io.Copy(out, r)
	err = out.Close()
	err = f.Close()
	if err != nil {
		mensaBotLog.Error("Error converting csv file to UTF-8")
	}
}

// Transforms the file from the uni-passau version of the csv file to a standard one
func transformFile(path string) {
	// Transforms csv file with separator ";" to a file with separator "," and also transforms all "," to ";"
	read, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	newContents := strings.Replace(string(read), ",", "*", -1)
	newContents = strings.Replace(newContents, ";", ",", -1)
	newContents = strings.Replace(newContents, "*", ";", -1)

	err = ioutil.WriteFile(path, []byte(newContents), 0)
	if err != nil {
		panic(err)
	}
}

// DownloadFile downloads the newest file based on the week number
func downloadFile(week string) error {
	// Downloads the csv file
	s1 := []string{week, ".csv"}
	filepath := strings.Join(s1, "")
	s2 := []string{"https://www.stwno.de/infomax/daten-extern/csv/UNI-P/", week, ".csv"}
	url := strings.Join(s2, "")

	// Create the file
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	mensaBotLog.Info("Finished Downloading")
	isoToUTF(filepath)
	transformFile(filepath)
	// Delete temporary file
	err = os.Remove(filepath + ".tmp")
	if err != nil {
		mensaBotLog.Error("Error removing old file: ", err)
	}
	mensaBotLog.Info("File Transformed")
	initArray()
	err = out.Close()
	checkmsg("Error closing body for creting file: ", err)
	err = resp.Body.Close()
	checkmsg("Error closing body for downloading file: ", err)
	return nil
}

// Open file for this week
func openactFile() *os.File {
	loc, _ := time.LoadLocation("Europe/Berlin")
	_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
	weekstring := strconv.Itoa(thisWeek)
	f, _ := os.Open(weekstring + ".csv")
	return f
}

// Update the food csv file
func updateFile() {
	// checks if new file has to be downloaded and does so - does also remove the old file
	loc, _ := time.LoadLocation("Europe/Berlin")
	_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
	weekstring := strconv.Itoa(thisWeek)
	if _, err := os.Stat(weekstring + ".csv"); os.IsNotExist(err) {

		// No actual file found
		mensaBotLog.Warning("No File for this week found - starting download --- ")
		err := downloadFile(weekstring)
		if err != nil {
			panic(err)
		}
		_, thisWeek := time.Now().In(loc).UTC().ISOWeek()
		prevweekstring := strconv.Itoa(thisWeek - 1)
		if _, err = os.Stat(prevweekstring + ".csv"); err == nil {
			err = os.Remove(prevweekstring + ".csv")
			if err != nil {
				mensaBotLog.Error("Error deleting Old File")
			}
		} else {
			mensaBotLog.Warning("No File from previous week to delete")
		}
	}
}

// Stop the program and kill hanging routines
func exit(quit chan bool) {
	// function for normal exit
	quit <- true
	simpleExit()
}

// Exit while ignoring running routines
func simpleExit() {
	// Exit without using graceful shutdown channels
	mensaBotLog.Info("Shutting down...")
	os.Exit(0)
}

// Print an error with given message
func checkmsg(message string, e error) {
	if e != nil {
		mensaBotLog.Fatal(message, e)
	}
}

// Print info regarding a given message
func printInfo(m *tb.Message) {
	mensaBotLog.Info("[UniPassauBot] " + m.Sender.Username + " - " + m.Sender.FirstName + " " + m.Sender.LastName + " - ID: " + strconv.Itoa(m.Sender.ID) + "Message: " + m.Text + "\n")
}

// Answer wrapper
func printAnswer(input string) {
	mensaBotLog.Info("[UniPassauBot] Answer: " + input)
}

// This is a workaround until the bot is fully integrated into systems over package borders.
func get(key string) string {
	switch key {
	case "isCorona":
		return "false"
	}
	return ""
}
