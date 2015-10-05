package ln

import (
	"fmt"
	"log/syslog"
	"sync"

	bsyslog "github.com/sendgrid/ln/vendor/blackjack"
)

// serialBlackjackSyslog funnels all logging calls through a single channel, to
// ensure that the CGO calls to syslog to not cause unbounded native thread
// growth.  If you are implementing graceful shutdown behavior in your app, be
// sure to call Close() before you exit to ensure that all pending messages are
// written out.
type serialBlackjackSyslog struct {
	msgChan chan serialBlackjackSyslogMessage
	wg      sync.WaitGroup

	// Ensure that we only close msgChan once
	closeChanOnce sync.Once
}

// serialBlackjackSyslogMessage encapsulates the message that is sent on the
// channel to the single goroutine writing syslog messages
type serialBlackjackSyslogMessage struct {
	priority syslog.Priority
	message  string
}

//calls Openlog with facility and tag. In the blackjack library, this parameter is incorrectly named
//as priority when it should actually be facility because linux openlog takes facility only
func newSerialBSyslogger(facility bsyslog.Priority, tag string) syslogger {
	bsyslog.Openlog(fmt.Sprintf("%s: %s", tag, tag), bsyslog.LOG_PID|bsyslog.LOG_CONS, facility)

	bsyslog := &serialBlackjackSyslog{
		msgChan: make(chan serialBlackjackSyslogMessage),
	}

	go bsyslog.workerLoop()

	return bsyslog
}

// workerLoop pulls messages off the channel and sends them to syslog
func (s *serialBlackjackSyslog) workerLoop() {
	for msg := range s.msgChan {
		switch msg.priority {
		case syslog.LOG_EMERG:
			bsyslog.Emerg(msg.message)
		case syslog.LOG_ALERT:
			bsyslog.Alert(msg.message)
		case syslog.LOG_CRIT:
			bsyslog.Crit(msg.message)
		case syslog.LOG_ERR:
			bsyslog.Err(msg.message)
		case syslog.LOG_NOTICE:
			bsyslog.Notice(msg.message)
		case syslog.LOG_WARNING:
			bsyslog.Warning(msg.message)
		case syslog.LOG_INFO:
			bsyslog.Info(msg.message)
		case syslog.LOG_DEBUG:
			bsyslog.Debug(msg.message)
		default:
			panic(fmt.Sprintf("ln/serialBlackjackSyslog: unrecognized severity %q", msg.priority))
		}

		// The corresponding Add() is called from each log method before it
		// sends the message on the channel)
		s.wg.Done()
	}
}

// Close will block until all pending messages are written out
func (s *serialBlackjackSyslog) Close() {
	// If there is a message that will be sent over the channel and has not been
	// logged yet, wait for it to be logged before we close
	s.wg.Wait()

	s.closeChanOnce.Do(func() {
		close(s.msgChan)
	})

	bsyslog.Closelog()
}

func (b *serialBlackjackSyslog) Emerg(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_EMERG, msg}
	return nil
}

func (b *serialBlackjackSyslog) Alert(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_ALERT, msg}
	return nil
}

func (b *serialBlackjackSyslog) Crit(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_CRIT, msg}
	return nil
}

func (b *serialBlackjackSyslog) Err(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_ERR, msg}
	return nil
}

func (b *serialBlackjackSyslog) Notice(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_NOTICE, msg}
	return nil
}

func (b *serialBlackjackSyslog) Warning(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_WARNING, msg}
	return nil
}

func (b *serialBlackjackSyslog) Info(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_INFO, msg}
	return nil
}

func (b *serialBlackjackSyslog) Debug(msg string) error {
	b.wg.Add(1)
	b.msgChan <- serialBlackjackSyslogMessage{syslog.LOG_DEBUG, msg}
	return nil
}
