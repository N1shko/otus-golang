package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	first *ListItem
	last  *ListItem
	len   int
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.first
}

func (l *list) Back() *ListItem {
	return l.last
}

func (l *list) PushFront(i interface{}) (item *ListItem) {
	item = &ListItem{Value: i, Prev: nil, Next: l.Front()}
	if l.Front() != nil {
		l.Front().Prev = item
	} else {
		l.last = item
	}
	l.first = item
	l.len++
	return
}

func (l *list) PushBack(i interface{}) (item *ListItem) {
	item = &ListItem{Value: i, Prev: l.Back(), Next: nil}
	if l.Back() != nil {
		l.Back().Next = item
	} else {
		l.first = item
	}
	l.last = item
	l.len++
	return
}

func (l *list) Remove(i *ListItem) {
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.last = i.Prev
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.first = i.Next
	}
	i = nil
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i != l.Front() {
		if i.Next != nil {
			i.Next.Prev = i.Prev
		} else {
			l.last = i.Prev
		}

		if i.Prev != nil {
			i.Prev.Next = i.Next
		}

		l.Front().Prev = i
		i.Next = l.Front()
		i.Prev = nil
		l.first = i
	}
}

func NewList() List {
	newList := list{first: nil, last: nil, len: 0}
	return &newList
}
