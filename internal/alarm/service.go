package alarm

type AlarmService struct {
	repo AlarmRepository
}

func NewAlarmService(repo AlarmRepository) *AlarmService {
	return &AlarmService{repo: repo}
}

func (s *AlarmService) GetAlarm(id string) (*Alarm, error) {
	return s.repo.GetAlarm(id)
}

func (s *AlarmService) GetAlarms() ([]*Alarm, error) {
	return s.repo.GetAlarms()
}

func (s *AlarmService) SetAlarm(alarm *Alarm) error {
	return s.repo.SetAlarm(alarm)
}

func (s *AlarmService) DeleteAlarm(id string) error {
	return s.repo.DeleteAlarm(id)
}
