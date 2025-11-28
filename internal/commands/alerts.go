package commands

import (
	"strconv"
	"strings"

	"teneo-agent-demo1/internal/services"
)

func CmdAlertAdd(alertSvc *services.AlertService, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: alert [symbol] [operator] [value]\nExample: alert btc > 65000", nil
	}

	symbol := args[0]
	var op, valStr string

	if len(args) == 2 {
		s := strings.TrimSpace(args[1])
		if strings.HasPrefix(s, ">=") || strings.HasPrefix(s, "<=") {
			op = s[:2]
			valStr = strings.TrimSpace(s[2:])
		} else if strings.HasPrefix(s, ">") || strings.HasPrefix(s, "<") || strings.HasPrefix(s, "=") {
			op = s[:1]
			valStr = strings.TrimSpace(s[1:])
		} else {
			return "Invalid format. Example: alert btc > 65000", nil
		}
	} else {
		op = strings.TrimSpace(args[1])
		valStr = strings.TrimSpace(args[2])
	}

	msg, err := alertSvc.Add(symbol, op, valStr)
	return msg, err
}

func CmdAlertList(alertSvc *services.AlertService) (string, error) {
	return alertSvc.List(), nil
}

func CmdAlertRemove(alertSvc *services.AlertService, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: alert_remove [index]", nil
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil || idx < 0 {
		return "Invalid index.", nil
	}
	return alertSvc.Remove(idx), nil
}

func CmdAlertClear(alertSvc *services.AlertService) (string, error) {
	return alertSvc.Clear(), nil
}
