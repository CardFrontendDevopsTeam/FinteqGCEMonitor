package cutofftimes

import (
	"github.com/tebeka/selenium"
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"log"
	"bytes"
	selenium2 "github.com/CardFrontendDevopsTeam/FinteqGCEMonitor/selenium"
	"errors"
)

type Service interface {
	DoCheck()
	parseInwardCutttoffTimes(i string)
}

type service struct {
	store    Store
	selenium selenium2.Service
	inward bool
}

func NewService(store Store, selenium selenium2.Service, inward bool) Service {
	return &service{store, selenium, inward}
}

var sodOk = map[string]struct{}{"SOD : ACK RECEIVED": {}}
var eodOk = map[string]struct{}{"EOD : ACK RECEIVED": {}}


func (s *service) DoCheck() {
	v , err := s.getData()
	if err != nil {
		s.selenium.HandleSeleniumError(err)
		return
	}
	var e []string
	for _, x := range v {
		if s.store.cutoffExists(x.service+x.subservice, x.destination) {
			if s.store.isInStartOfDay(x.service+x.subservice, x.destination) {
				_, ok := sodOk[x.status]
				if !ok {
					e = append(e, fmt.Sprintf("invalid status for service %v, sub service %v, status %v", x.service+x.subservice, x.destination, x.status))
				}
			}
		} else {
			log.Printf("No records found for service %v, subservice %v", x.service+x.subservice, x.destination)
		}
	}
	if len(e) > 0 {
		b := bytes.Buffer{}
		for _, s := range e {
			b.WriteString(s)
			b.WriteString("\n")
		}
		s.selenium.HandleSeleniumError(&selenium2.SeleniumnError{false, errors.New(b.String())})
	}
}


func (s service)parseInwardCutttoffTimes(i string) {
	scanner := bufio.NewScanner(strings.NewReader(i))

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Parsing Line: %v", line)
		tokens := strings.Split(line, ";")
		sodHour := tokens[2]
		sodHour = strings.TrimSpace(sodHour)
		if len(sodHour) == 0 {
			continue
		}
		s.parseBlock(tokens[0], tokens[1], tokens[2], tokens[3], tokens[4])
		s.parseBlock(tokens[0], tokens[1], tokens[5], tokens[6], tokens[7])
	}

}

func (s service) parseBlock(service, subservice, sodTime, eodTime, days string) {
	sodTime = strings.TrimSpace(sodTime)
	eodTime = strings.TrimSpace(eodTime)

	if len(sodTime) == 0 {
		return
	}

	sodTime = strings.Replace(sodTime, "A ", "", 1)
	eodTime = strings.Replace(eodTime, "A ", "", 1)

	days = strings.TrimSpace(days)
	days = strings.Replace(days, "(ph)", "", -1)

	c := cutoffTime{Service: service, SubService: subservice}

	sod := strings.Split(sodTime, "H")
	c.SodHour, _ = strconv.Atoi(sod[0])
	c.SodMinute, _ = strconv.Atoi(sod[1])

	eod := strings.Split(eodTime, "H")
	c.EodHour, _ = strconv.Atoi(eod[0])
	c.EodMinute, _ = strconv.Atoi(eod[1])

	if days == "Mon - Sun" {
		i := 0
		for i < 7 {
			c.DayOfWeek = i
			s.store.saveCutoff(c)
			i++
		}
		return
	}

	if days == "Mon - Fri" {
		i := 1
		for i < 6 {
			c.DayOfWeek = i
			s.store.saveCutoff(c)
			i++
		}
		return
	}

	//If its not Monday - Sunday or Monday to Friday, it must be Sat - Sun

	c.DayOfWeek = 6
	s.store.saveCutoff(c)

	c.DayOfWeek = 0
	s.store.saveCutoff(c)
}

func (s service) getData() ([]inwardService, error) {

	err := s.selenium.Driver().Wait(func(wb selenium.WebDriver) (bool, error) {
		elem, err := wb.FindElement(selenium.ByPartialLinkText, "Service Options")
		if err != nil {
			return false, nil
		}
		return elem.IsDisplayed()
	})

	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}
	}

	elem, err := s.selenium.Driver().FindElement(selenium.ByPartialLinkText, "Service Options")
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	err = elem.Click()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	err = s.selenium.WaitForWaitFor()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	link := ""
	if s.inward {
		link = "INWARD SERVICE OPTIONS"

	} else {
		link = "OUTWARD SERVICE OPTIONS"
	}

	s.selenium.Driver().Wait(func(wb selenium.WebDriver) (bool, error) {
		elem, err := wb.FindElement(selenium.ByPartialLinkText, link)
		if err != nil {
			return false, nil
		}
		return elem.IsDisplayed()
	})

	elem, err = s.selenium.Driver().FindElement(selenium.ByPartialLinkText, link)
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	err = elem.Click()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	err = s.selenium.WaitForWaitFor()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	v := s.checkTable()
	elem, err = s.selenium.Driver().FindElement(selenium.ByPartialLinkText, "2")
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}
	err = elem.Click()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	err = s.selenium.WaitForWaitFor()
	if err != nil {
		return nil, &selenium2.SeleniumnError{true, err}

	}

	b := s.checkTable()
	for _, x := range b {
		v = append(v, x)
	}
	return v, nil
}

func (s *service)checkTable() []inwardService {

	table := ""

	if s.inward {
		table = "TABLEINWARDSERVICES"

	} else {
		table = "TABLEOUTWARDSERVICES"
	}

	var v []inwardService
	s.selenium.Driver().Wait(func(wb selenium.WebDriver) (bool, error) {
		elem, err := wb.FindElement(selenium.ByXPATH, "//table[@id='"+table+"']/tbody/tr[1]/td[13]")
		if err != nil {
			return false, nil
		}
		return elem.IsDisplayed()
	})
	i := 1
	for i < 50 {
		var service string

		service, err := s.getTableElement(i, 2, table)
		if err != nil {
			return v
		}
		subService, err := s.getTableElement(i, 3, table)
		if err != nil {
			return v
		}
		destinationCode, err := s.getTableElement(i, 4, table)
		if err != nil {
			return v
		}
		status, err := s.getTableElement(i, 13, table)
		if err != nil {
			return v
		}
		v = append(v, inwardService{service: service, destination: destinationCode, subservice: subService, status: status})
		i++
	}
	return v
}

func (s *service)getTableElement(row, column int, table string) (string, error) {
	elem, err := s.selenium.Driver().FindElement(selenium.ByXPATH, fmt.Sprintf("//table[@id='"+table+"']/tbody/tr[%v]/td[%v]", row, column))
	if err != nil {
		return "", err
	}
	return elem.Text()
}

type inwardService struct {
	service, subservice, destination, status string
}