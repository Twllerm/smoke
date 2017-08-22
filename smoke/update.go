package smoke

import (
	"time"
	"strconv"
	"log"
	"bot/bot"
	"sort"
)

func (s *Smoke) update() {
	log.Println("Smoke::update START")
	s.updateWithNotify("", 0)
	log.Println("Smoke::update END")
}

func (s *Smoke) updateWithNotify(msg string, omitChatId int) {
	log.Println("Smoke::updateWithNotify START")

	for _, sc := range s.SCs {
		if sc.Locked {
			continue
		}

		r := sc.PostResponse
		r.Text = s.Format()
		go sc.Context.Send(r)
		if msg != "" {
			if sc.Account.ChatId != omitChatId {
				go s.notifyOne(msg, sc)
			}
		}
	}
	log.Println("Smoke::updateWithNotify END")
}

func (s *Smoke) Format() string {
	log.Println("Smoke::Format START")
	var when string
	if s.min < 1 {
		when = "сейчас"
	} else {
		when = "через *" + strconv.Itoa(s.min) + "* минут"
	}

	res := "*" + s.getUniqueUserName(s.CreatorSC.Account) + "* из группы *" +
		s.group.Name + "*" + " вызывает " + when + "\n\n"

	var keys []int
	for chatId := range s.SCs {
		keys = append(keys, chatId)
	}

	sort.Ints(keys)

	for _, chatId := range keys {
		sc := s.SCs[chatId]
		res += s.answer(sc)
		res += s.comment(sc)
		res += "\n"
	}

	res += "\n_Ответьте на это сообщение для комментария_"
	log.Println("Smoke::Format END")
	return res
}

func (s *Smoke) answer(sc *SmokerContext) string {
	if sc.Answered {
		return s.getUniqueUserName(sc.Account) + " - " + boolToAnswer(sc.Going)
	}
	return s.getUniqueUserName(sc.Account) + " - "
}

func (s *Smoke) comment(sc *SmokerContext) string {
	if sc.Comment != "" {
		return ", _" + sc.Comment + "_ "
	}
	return ""
}

func (s *Smoke) notifyOne(msg string, smokerContext *SmokerContext) {
	log.Println("Smoke::notifyOne START")

	if !s.SCs[smokerContext.Account.ChatId].Going {
		return
	}

	r := &bot.Response{
		Text: msg,
	}

	smokerContext.Context.Send(r)
	time.Sleep(15 * time.Second)
	smokerContext.Context.DeleteResponse(r)
	log.Println("Smoke::notifyOne END")
}

func (s *Smoke) goingSmokers() int {
	log.Println("Smoke::goingSmokers START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	goingSmokers := 0
	for _, sc := range s.SCs {
		if sc.Going {
			goingSmokers++
		}
	}
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	log.Println("Smoke::goingSmokers END")
	return goingSmokers
}

func boolToAnswer(going bool) string {
	if going {
		return "Да"
	}
	return "Нет"
}

func (s *Smoke) decrement() {
	log.Println("Smoke:lock")
	s.lock.Lock()
	log.Println("decrementing min")
	s.min--
	log.Println("Smoke:unlock")
	s.lock.Unlock()
}

func (s *Smoke) notifyAll(msg string) {
	s.notifyAllExcept(msg, 0)
}

func (s *Smoke) notifyAllExcept(msg string, omitChatId int) {
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId || !smokerContext.Going {
			continue
		}
		go s.notifyOne(msg, smokerContext)
	}
}

func (s *Smoke) delayedCancel(min int) {
	log.Println("Smoke::delayedCancel START")
	s.delayedCancelEnabled = true
	defer func() {
		s.delayedCancelEnabled = false
	}()
	t := time.NewTicker(time.Duration(min) * time.Minute)
	select{
		case <-t.C:
			log.Println("Smoke::delayedCancel END")
			s.Cancel(false)
		case <-s.cancelDelayedCancel:
			log.Println("Smoke::delayedCancel END. cancelLifecycle")
			return
	}
}
