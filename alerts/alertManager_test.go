package alerts

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/magmasystems/SlackStockSlashCommand/stockbot"
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
)

func TestCreateAlertManager(t *testing.T) {
	type args struct {
		bot *stockbot.Stockbot
	}
	tests := []struct {
		name string
		args args
		want *AlertManager
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAlertManager(tt.args.bot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAlertManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_Dispose(t *testing.T) {
	tests := []struct {
		name         string
		alertManager *AlertManager
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alertManager.Dispose()
		})
	}
}

func Test_getDbConnectionInfo(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDbConnectionInfo(); got != tt.want {
				t.Errorf("getDbConnectionInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_HandleQuoteAlert(t *testing.T) {
	type args struct {
		slashCommand slack.SlashCommand
		writer       http.ResponseWriter
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alertManager.HandleQuoteAlert(tt.args.slashCommand, tt.args.writer)
		})
	}
}

func TestAlertManager_listAllAlerts(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		want         string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.listAllAlerts(tt.args.userID); got != tt.want {
				t.Errorf("AlertManager.listAllAlerts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_getAlert(t *testing.T) {
	type args struct {
		userID string
		params *createAlertParams
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		want         *quoteAlert
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.getAlert(tt.args.userID, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AlertManager.getAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_insertNewAlert(t *testing.T) {
	type args struct {
		userID string
		params *createAlertParams
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		want         string
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.alertManager.insertNewAlert(tt.args.userID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlertManager.insertNewAlert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AlertManager.insertNewAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_setWasNotified(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alertManager.setWasNotified(tt.args.id)
		})
	}
}

func TestAlertManager_deleteAllAlerts(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alertManager.deleteAllAlerts(tt.args.userID)
		})
	}
}

func TestAlertManager_deleteAlert(t *testing.T) {
	type args struct {
		userID string
		params *createAlertParams
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		want         string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.deleteAlert(tt.args.userID, tt.args.params); got != tt.want {
				t.Errorf("AlertManager.deleteAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_CheckForPriceBreaches(t *testing.T) {
	type args struct {
		stockbot *stockbot.Stockbot
		callback func(PriceBreachNotification)
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alertManager.CheckForPriceBreaches(tt.args.stockbot, tt.args.callback)
		})
	}
}

func TestAlertManager_GetQuotesForAlerts(t *testing.T) {
	type args struct {
		stockbot *stockbot.Stockbot
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		want         []PriceInfo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.GetQuotesForAlerts(tt.args.stockbot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AlertManager.GetQuotesForAlerts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_GetAlertedSymbols(t *testing.T) {
	tests := []struct {
		name         string
		alertManager *AlertManager
		want         []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.GetAlertedSymbols(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AlertManager.GetAlertedSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertManager_SavePrices(t *testing.T) {
	type args struct {
		prices []PriceInfo
	}
	tests := []struct {
		name         string
		alertManager *AlertManager
		args         args
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.alertManager.SavePrices(tt.args.prices); (err != nil) != tt.wantErr {
				t.Errorf("AlertManager.SavePrices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlertManager_GetPriceBreaches(t *testing.T) {
	tests := []struct {
		name         string
		alertManager *AlertManager
		want         []PriceBreachNotification
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alertManager.GetPriceBreaches(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AlertManager.GetPriceBreaches() = %v, want %v", got, tt.want)
			}
		})
	}
}
