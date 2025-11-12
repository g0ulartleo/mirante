package signal

import "time"

type Service struct {
	repo SignalRepository
}

func NewService(repo SignalRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WriteSignal(signal Signal) error {
	return s.repo.Save(signal)
}

func (s *Service) GetAlarmLatestSignals(alarmID string, limit int) ([]Signal, error) {
	return s.repo.GetAlarmLatestSignals(alarmID, limit)
}

func (s *Service) AlarmHasChangedStatus(alarmID string) (bool, error) {
	signals, err := s.GetAlarmLatestSignals(alarmID, 2)
	if err != nil {
		return false, err
	}
	if len(signals) == 0 {
		return false, nil
	}
	if len(signals) == 1 {
		return true, nil
	}
	return signals[0].Status != signals[1].Status, nil
}

func (s *Service) GetAlarmHealth(alarmID string) (Status, error) {
	return s.repo.GetAlarmHealth(alarmID)
}

func (s *Service) CleanOldSignals() error {
	return s.repo.CleanOldSignals()
}

func (s *Service) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	return s.repo.CountUnhealthySince(alarmID, since)
}
