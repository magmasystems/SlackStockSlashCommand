package alerts

import (
	sql "database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	config "github.com/magmasystems/SlackStockSlashCommand/configuration"
	fr "github.com/magmasystems/SlackStockSlashCommand/framework"
	logging "github.com/magmasystems/SlackStockSlashCommand/framework/logging"
	slackmessaging "github.com/magmasystems/SlackStockSlashCommand/slackmessaging"
	stockbot "github.com/magmasystems/SlackStockSlashCommand/stockbot"

	// Need this for postgres
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
)

// AlertManager - handles all alerting
type AlertManager struct {
	fr.Disposable
	AlertManagerOps
	db       *sql.DB
	config   *config.AppSettings
	stockBot *stockbot.Stockbot
}

type quoteAlert struct {
	id            int
	slackUserName string
	channel       string
	symbol        string
	price         float64
	direction     string
	wasNotified   bool
}

type createAlertParams struct {
	channel   string
	symbol    string
	price     float64
	direction string
	delete    bool
	deleteAll bool
}

// PriceInfo - holds the price for a stock
type PriceInfo struct {
	Symbol string
	Price  float64
}

// PriceBreachNotification - a notification that a price limit has been breached
type PriceBreachNotification struct {
	SubscriptionID int
	SlackUserName  string
	Channel        string
	Symbol         string
	TargetPrice    float64
	CurrentPrice   float64
	Direction      string
}

// AlertManagerOps - defines all operationsn that the AlertManager can do
type AlertManagerOps interface {
	CheckForPriceBreaches(stockbot *stockbot.Stockbot, callback func(PriceBreachNotification))
	HandleQuoteAlert(slashCommand slack.SlashCommand, w http.ResponseWriter)
	GetAlertedSymbols() []string
	GetPriceBreaches() []PriceBreachNotification
	GetQuotesForAlerts(stockbot *stockbot.Stockbot) []PriceInfo
	SavePrices(prices []PriceInfo) error
}

// CreateAlertManager - creates and initializes a new AlertManager
func CreateAlertManager(bot *stockbot.Stockbot) *AlertManager {
	alertManager := new(AlertManager)

	db, err := sql.Open("postgres", getDbConnectionInfo())
	if err != nil {
		logging.Infoln("Alert Manager: cannot open the postgres database")
		logging.Infoln(fmt.Sprint(err))
		time.Sleep(time.Duration(1) * time.Second)
		logging.Panic(err)
	}
	logging.Infoln("Alert Manager: was able to open the database.")

	/*
		logging.Infoln("Alert Manager: About to ping the database")
		err = db.Ping()
		if err != nil {
			logging.Infoln("Alert Manager: cannot ping the postgres database")
			logging.Infoln(fmt.Sprint(err))
			logging.Panic(err)
		}
		logging.Infoln("Alert Manager: was able to ping the database")
	*/

	alertManager.db = db
	alertManager.stockBot = bot

	configMgr := new(config.ConfigManager)
	alertManager.config = configMgr.Config()
	logging.Infoln("Alert Manager: fetched the config info")

	logging.Infoln("Alert Manager: returning from creating the alert manager")
	return alertManager
}

// Dispose - clean up resources
func (alertManager *AlertManager) Dispose() {
	if alertManager.db != nil {
		alertManager.db.Close()
		alertManager.db = nil
	}
}

func getDbConnectionInfo() string {
	configMgr := new(config.ConfigManager)
	appSettings := configMgr.Config()

	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s", appSettings.Database.Host, appSettings.Database.Port, appSettings.Database.DbName)
	if appSettings.Database.User != "" {
		psqlInfo += fmt.Sprintf(" user=%s", appSettings.Database.User)
	}
	if appSettings.Database.Password != "" {
		psqlInfo += fmt.Sprintf(" password=%s", appSettings.Database.Password)
	}
	if appSettings.Database.SSL {
		psqlInfo += fmt.Sprintf(" sslmode=require")
	}

	// [host=slackstockbot.cfaf3ksbohge.us-east-2.rds.amazonaws.com port=5432 dbname=slackstockbot sslmode=disable user=magmasystems password=magma123]
	logging.Infof("Alert Manager: The connection info is [%s]\n", psqlInfo)

	return psqlInfo
}

// HandleQuoteAlert - parses and dispatches a /quote-alert command from Slack
func (alertManager *AlertManager) HandleQuoteAlert(slashCommand slack.SlashCommand, writer http.ResponseWriter) {
	outputText := ""
	args := strings.Split(strings.Trim(slashCommand.Text, " "), " ")
	logging.Infof("Alert Manager: Got new Quote Alert with the args [%s]\n", fmt.Sprint(args))

	var params *createAlertParams

	// If no args were passed, then we just send back a list of all of the alerts that a user has
	if args[0] != "" {
		// syntax: /quotealert symbol price [below] [delete]
		params = &createAlertParams{"", "", 0.0, "ABOVE", false, false}
		for i := 0; i < len(args); i = i + 1 {
			param := strings.ToLower(args[i])

			if param == "deleteall" {
				params.deleteAll = true
			} else if param == "delete" {
				params.delete = true
			} else if param == "above" || param == "below" {
				params.direction = strings.ToUpper(param)
			} else if strings.ContainsAny(args[i], "0123456789") {
				params.price, _ = strconv.ParseFloat(param, 32)
			} else if param[0] == '#' {
				params.channel = args[i] // if the arg starts with #, assume it is the name of a channel to send the notification to
			} else {
				params.symbol = param
			}
		}
	}

	logging.Infof("Alert Manager: the alert params are [%s]\n", fmt.Sprint(params))

	if params != nil {
		if params.deleteAll {
			alertManager.deleteAllAlerts(slashCommand.UserID)
			outputText = fmt.Sprintf("All alerts deleted for user %s", slashCommand.UserName)
		} else if params.delete {
			alertManager.deleteAlert(slashCommand.UserID, params)
			outputText = fmt.Sprintf("Alert deleted for user %s", slashCommand.UserName)
		} else {
			newID, err := alertManager.insertNewAlert(slashCommand.UserID, params)
			if err != nil {
				outputText = err.Error() // maybe the user request a symbol that is not a stock
			} else {
				outputText = fmt.Sprintf("Alert %s Created for user %s", newID, slashCommand.UserName)
			}
		}
	} else {
		outputText = alertManager.listAllAlerts(slashCommand.UserID)
	}

	// Send the response back to Slack
	logging.Infoln(outputText)
	slackmessaging.WriteResponse(writer, outputText)
}

func (alertManager *AlertManager) listAllAlerts(userID string) string {
	q := new(quoteAlert)
	outputText := ""

	sqlStatement := `SELECT id, slackuser, channel, symbol, targetprice, wasnotified, direction
	FROM slackstockbot.alertsubscription
	WHERE slackuser = $1`

	rows, err := alertManager.db.Query(sqlStatement, userID)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&q.id, &q.slackUserName, &q.channel, &q.symbol, &q.price, &q.wasNotified, &q.direction)
		if err != nil {
			panic(err)
		}
		outputText += fmt.Sprintf("%s\t%3.2f (%s)\n", q.symbol, q.price, q.direction)
	}

	return outputText
}

func (alertManager *AlertManager) getAlert(userID string, params *createAlertParams) *quoteAlert {
	sqlStatement := `SELECT id, slackuser, channel, symbol, targetprice, wasnotified, direction
	FROM slackstockbot.alertsubscription
	WHERE slackuser = $1 AND symbol = $2 AND direction = $3`

	logging.Infof("AlertManager.getAlert: %s\n", sqlStatement)
	row := alertManager.db.QueryRow(sqlStatement, userID, params.symbol, params.direction)
	logging.Infof("AlertManager.getAlert: returned with row [%+v]\n", row)

	q := new(quoteAlert)

	switch err := row.Scan(&q.id, &q.slackUserName, &q.channel, &q.symbol, &q.price, &q.wasNotified, &q.direction); err {
	case sql.ErrNoRows:
		logging.Infoln("AlertManager.getAlert: no rows returned. Returning nil")
		return nil
	case nil:
		logging.Infoln("AlertManager.getAlert: no errors returned. Returning the alert")
		return q
	default:
		logging.Fatal(err)
		panic(err)
	}
}

func (alertManager *AlertManager) insertNewAlert(userID string, params *createAlertParams) (string, error) {
	logging.Infof("AlertManager.insertNewAlert: Getting alerts for [userID %s, symbol %s]", userID, params.symbol)
	quoteAlert := alertManager.getAlert(userID, params)
	logging.Infof("AlertManager.insertNewAlert: Returned with quoteAlert %+v", quoteAlert)

	if quoteAlert != nil {
		// The record already exists. Just update the fields
		sqlStatement := `UPDATE slackstockbot.alertsubscription SET targetprice = $1, direction = $2 WHERE id = $3`
		res, err := alertManager.db.Exec(sqlStatement, params.price, params.direction, quoteAlert.id)
		if err != nil {
			logging.Fatal(err)
			panic(err)
		}

		rowsUpdated, _ := res.RowsAffected()
		if rowsUpdated == 1 {
			return strconv.Itoa(quoteAlert.id), nil
		}
		return "0", nil
	}

	// See if the symbol is a valid stock by trying to fetch the current price
	quoteInfo := alertManager.stockBot.QuoteSingle(params.symbol)
	if len(quoteInfo) == 0 || quoteInfo[0].LastPrice == 0 {
		return "", errors.New("The symbol is invalid")
	}

	// https://www.calhoun.io/inserting-records-into-a-postgresql-database-with-gos-database-sql-package/
	sqlStatement := `
INSERT INTO slackstockbot.alertsubscription (slackuser, channel, symbol, targetprice, wasnotified, direction)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id`
	id := 0
	logging.Infof("AlertManager.insertNewAlert: %s\n", sqlStatement)
	err := alertManager.db.QueryRow(sqlStatement, userID, params.channel, strings.ToUpper(params.symbol), params.price, false, params.direction).Scan(&id)
	logging.Infof("AlertManager.insertNewAlert: QueryRow returned with err [%s]\n", fmt.Sprint(err))
	if err != nil {
		logging.Panic(err)
	}

	logging.Infof("AlertManager.insertNewAlert: returning id %d\n", id)
	return strconv.Itoa(id), nil
}

func (alertManager *AlertManager) setWasNotified(id int) {
	sqlStatement := `UPDATE slackstockbot.alertsubscription SET wasnotified = true WHERE id = $1`
	alertManager.db.Exec(sqlStatement, id)
}

func (alertManager *AlertManager) deleteAllAlerts(userID string) {
	sqlStatement := `DELETE FROM slackstockbot.alertsubscription WHERE slackuser = $1;`
	_, err := alertManager.db.Exec(sqlStatement, userID)
	if err != nil {
		panic(err)
	}
}

func (alertManager *AlertManager) deleteAlert(userID string, params *createAlertParams) string {
	sqlStatement := `DELETE FROM slackstockbot.alertsubscription WHERE slackuser = $1 AND symbol = $2;`
	_, err := alertManager.db.Exec(sqlStatement, userID, params.symbol)
	if err != nil {
		panic(err)
	}
	return "alert deleted"
}

// CheckForPriceBreaches - gets called by the application at periodic intervals to check for price breaches
func (alertManager *AlertManager) CheckForPriceBreaches(stockbot *stockbot.Stockbot, callback func(PriceBreachNotification)) {
	// Get the latest quotes
	prices := alertManager.GetQuotesForAlerts(stockbot)
	if prices == nil {
		return
	}

	// Save the prices to the database
	alertManager.SavePrices(prices)

	// Check for any price breaches
	notifications := alertManager.GetPriceBreaches()

	// Go through all of the price breaches and notify the Slack user
	for _, notification := range notifications {
		// Set the wasNotified field to TRUE on the alert
		alertManager.setWasNotified(notification.SubscriptionID)

		// Do the notification to slack synchronously
		callback(notification)
	}
}

// GetQuotesForAlerts - gets the current prices for all alertable stocks
func (alertManager *AlertManager) GetQuotesForAlerts(stockbot *stockbot.Stockbot) []PriceInfo {
	symbols := alertManager.GetAlertedSymbols()

	go func() {
		stockbot.QuoteAsync(symbols)
	}()

	select {
	case quotes := <-stockbot.QuoteReceived:
		var priceInfos []PriceInfo
		for _, q := range quotes {
			pi := PriceInfo{Symbol: strings.ToUpper(q.Symbol), Price: float64(q.LastPrice)}
			if pi.Price > 0 {
				priceInfos = append(priceInfos, pi)
			}
		}
		return priceInfos

	case <-time.After(10 * time.Second):
		return nil
	}
}

// GetAlertedSymbols - gets a unique list of all of the stocks, over all of the users, that we need prices for
func (alertManager *AlertManager) GetAlertedSymbols() []string {
	var symbols []string
	var symbol string

	sqlStatement := `SELECT DISTINCT symbol FROM slackstockbot.alertsubscription ORDER BY symbol;`

	rows, err := alertManager.db.Query(sqlStatement)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&symbol)
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
		symbols = append(symbols, symbol)
	}

	logging.Infoln("The alerted symbols are:")
	logging.Infoln(fmt.Sprint(symbols))

	return symbols
}

// SavePrices - saves a list of updated quotes in the database
func (alertManager *AlertManager) SavePrices(prices []PriceInfo) error {
	sqlStatement := `DELETE FROM slackstockbot.stockprice;`
	_, err := alertManager.db.Exec(sqlStatement)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	// Insert multiple values
	sqlStatement = "INSERT INTO slackstockbot.stockprice (symbol, price, time) VALUES "
	for _, pi := range prices {
		if pi.Price > 0 {
			sqlStatement += fmt.Sprintf("('%s', %3.2f, current_date),", strings.ToUpper(pi.Symbol), pi.Price)
		}
	}

	// Get rid of the trailing comma and append a semicolon to terminate the statement
	sqlStatement = strings.TrimRight(sqlStatement, ",") + ";"

	_, err = alertManager.db.Exec(sqlStatement)

	logging.Infoln("Saving the new prices to the database:")
	logging.Infoln(sqlStatement)

	return err
}

// GetPriceBreaches - compare all of the alerts in the database with the current prices, and returna list of breaches
func (alertManager *AlertManager) GetPriceBreaches() []PriceBreachNotification {
	var notifications []PriceBreachNotification
	q := PriceBreachNotification{}

	// This SQL will compare all of the alerts against the list of current quotes, and identify those alerts
	// which have price breaches in either direction.
	sqlStatement := `SELECT a.id, a.slackuser, a.channel, a.symbol, a.targetprice, a.direction, p.price 
	FROM slackstockbot.alertsubscription a, slackstockbot.stockprice p
	WHERE a.wasnotified = false AND a.symbol = p.symbol AND p.price > 0 AND
	      ( (a.direction = 'ABOVE' AND p.price >= a.targetprice) OR (a.direction = 'BELOW' AND p.price <= a.targetprice) );`

	rows, err := alertManager.db.Query(sqlStatement)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&q.SubscriptionID, &q.SlackUserName, &q.Channel, &q.Symbol, &q.TargetPrice, &q.Direction, &q.CurrentPrice)
		if err != nil {
			panic(err)
		}

		if q.CurrentPrice != 0 {
			notifications = append(notifications, q)
		}
	}

	return notifications
}
