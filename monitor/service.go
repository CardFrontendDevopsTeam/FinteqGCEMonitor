package monitor

import (
	"github.com/weAutomateEverything/finteqGCEMonitor/cutofftimes"
	"github.com/weAutomateEverything/finteqGCEMonitor/gceSelenium"
	"github.com/weAutomateEverything/finteqGCEMonitor/gceservices"
	"github.com/weAutomateEverything/go2hal/alert"
	"golang.org/x/net/context"
	"os"
	"time"
)

type Service interface {
}

type service struct {
	alert       alert.Service
	selenium    gceSelenium.Service
	gceService  gceservices.Service
	cutofftimes cutofftimes.Service
}

func NewService(alert alert.Service, selenium gceSelenium.Service, gceService gceservices.Service,
	cutofftimes cutofftimes.Service) Service {
	s := &service{alert, selenium, gceService, cutofftimes}
	go func() {
		s.startMonitor()
	}()

	return s

}

func (s *service) startMonitor() {
	for true {
		s.doCheck()
		time.Sleep(10 * time.Minute)

	}

}

func (s *service) doCheck() {

	err := s.selenium.NewClient()
	if err != nil {
		s.alert.SendError(context.TODO(), err)
		return
	}
	defer s.selenium.Driver().Quit()

	err = s.selenium.Driver().Get(endpoint())
	if err != nil {
		s.selenium.HandleSeleniumError(true, err)
		return
	}

	err = s.selenium.WaitForWaitFor()

	if err != nil {
		s.selenium.HandleSeleniumError(true, err)
		return
	}

	s.gceService.RunServiceCheck(true)
	s.gceService.RunServiceCheck(false)

	//s.cutofftimes.DoCheck(true)
	//s.cutofftimes.DoCheck(false)

}

func endpoint() string {
	return os.Getenv("gce_endpoint")
}
