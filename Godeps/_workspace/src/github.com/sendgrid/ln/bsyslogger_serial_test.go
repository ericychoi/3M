package ln

import (
	"testing"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

func TestSerialBlackjackLevels(t *testing.T) {

	levels := []string{"EMERG", "ALERT", "CRIT", "WARNING", "INFO", "DEBUG", "NOTICE", "ERR"}
	facilities := []string{"KERN", "USER", "MAIL", "DAEMON", "AUTH", "SYSLOG", "LPR", "NEWS", "UUCP", "CRON", "AUTHPRIV", "FTP", "LOCAL0", "LOCAL1", "LOCAL2", "LOCAL3", "LOCAL4", "LOCAL5", "LOCAL6", "LOCAL7"}

	var serialBlackjack LevelLogger

	for _, level := range levels {
		New("syslog", level, "LOCAL0", "tag")
	}
	for _, facility := range facilities {
		serialBlackjack = New("syslog", "DEBUG", facility, "tag")
	}

	// serialBlackjack.Emerg("foo", nil) // commented out so it doesn't spam/error out on chef runs, but it passes, really!
	serialBlackjack.Alert("foo", nil)
	serialBlackjack.Crit("foo", nil)
	serialBlackjack.Warning("foo", nil)
	serialBlackjack.Info("foo", nil)
	serialBlackjack.Debug("foo", nil)
	serialBlackjack.Notice("foo", nil)
	serialBlackjack.Err("foo", nil)
	serialBlackjack.Debug("foo", nil)
}

func BenchmarkSerialBlackjack(b *testing.B) {
	logger := newSerialBSyslogger(bsyslog.LOG_DEBUG, "BenchmarkSerialBlackjack")
	for i := 0; i < b.N; i++ {
		logger.Alert("Alert")
		logger.Crit("Crit")
		logger.Warning("Warning")
		logger.Info("Info")
		logger.Debug("Debug")
		logger.Notice("Notice")
		logger.Err("Err")
		logger.Debug("Debug")
	}
}
