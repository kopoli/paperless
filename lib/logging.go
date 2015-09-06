
// +build !windows

package paperless

import (
	"log"
	"log/syslog"

)

func SetupLogging()  {
	syslg, err := syslog.New(syslog.LOG_ERR|syslog.LOG_DAEMON, "paperless")
	if err != nil {
		log.Fatal("Creating a syslog logger failed")
	}
	log.SetOutput(syslg)
}
