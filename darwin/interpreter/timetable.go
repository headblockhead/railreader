package interpreter

import "github.com/headblockhead/railreader/darwin/unmarshaller"

func (u UnitOfWork) InterpretTimetable(timetable unmarshaller.Timetable) error {
	u.log.Debug("interpreting a Timetable")
	return nil // TODO
}
