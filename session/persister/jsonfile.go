package persister

import "OpenSplit/session"

type JsonFile struct {
	fileName string
}

func (j JsonFile) Save(id string, splitFile *session.SplitFile) error {
	//TODO implement me
	panic("implement me")
}

func (j JsonFile) Load(id string) (*session.SplitFile, error) {
	//TODO implement me
	panic("implement me")
}
