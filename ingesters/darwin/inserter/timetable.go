package inserter

import "github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"

func (u *UnitOfWork) InsertTimetable(timetable unmarshaller.Timetable, filename string) error {
	return nil
}

func (u *UnitOfWork) TimetableAlreadyImported(filename string) (bool, error) {
	return false, nil
}
