package lobby

import (
	"fmt"
	"sudojo/domain/sudoku"
)

type Move struct {
	Row    int   `json:"row"`
	Column int   `json:"column"`
	Value  int   `json:"value"`
	Time   int64 `json:"time"`
}

type Content struct {
	Name     string  `json:"name"`
	Points   []*Move `json:"points"`
	Mistakes []*Move `json:"mistakes"`
}

type Report struct {
	Id       string         `json:"id"`
	Created  int64          `json:"created_at"`
	Finished *int64         `json:"finished_at"`
	Board    *sudoku.Sudoku `json:"board"`
	Content  []*Content     `json:"content"`
}

func insert(array []*Move, item *Move) []*Move {
	index := 0
	for index < len(array) && array[index].Time <= item.Time {
		index++
	}

	array = append(array, nil)
	copy(array[index+1:], array[index:])
	array[index] = item

	return array
}

func (l *Lobby) Export() *Report {
	if l.Game.Finished == nil {
		return nil
	}

	<-l.logger.done // block on the logger until all messages are stored

	content := make(map[string]*Content)
	check := make(map[string]struct{})
	l.lock.RLock()
	for t, p := range l.players {
		content[t] = &Content{
			Name:     p.name,
			Points:   []*Move{},
			Mistakes: []*Move{},
		}
	}
	l.lock.RUnlock()

	l.logger.lock.RLock()
	for _, msg := range l.logger.storage {
		for key := range content {
			if msg.Player == key {
				if msg.Value == sudoku.EmptyCell {
					continue
				}

				if l.Game.Solution[msg.Row][msg.Column] != msg.Value {
					content[key].Mistakes = insert(content[key].Mistakes, &Move{
						Row:    msg.Row,
						Column: msg.Column,
						Value:  msg.Value,
						Time:   msg.Time,
					})
				} else {
					pos := fmt.Sprintf("%d%d", msg.Row, msg.Column)
					if _, exist := check[pos]; !exist {
						check[pos] = struct{}{}
						content[key].Points = insert(content[key].Points, &Move{
							Row:    msg.Row,
							Column: msg.Column,
							Value:  msg.Value,
							Time:   msg.Time,
						})
					}
				}
			}
		}
	}
	l.logger.lock.RUnlock()

	rep := &Report{
		Id:       l.Id,
		Created:  l.Game.Created,
		Finished: l.Game.Finished,
		Board:    l.Game.Solution,
	}
	for _, v := range content {
		rep.Content = append(rep.Content, v)
	}

	return rep
}
