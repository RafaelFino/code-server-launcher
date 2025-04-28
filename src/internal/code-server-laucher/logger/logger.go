package logger

import (
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New(){
	Out:   logrus.StandardLogger().Out,
	Level: logrus.DebugLevel,
	Formatter: &logrus.TextFormatter{
		//DisableColors: true,	
		FullTimestamp: true,
	},	
}