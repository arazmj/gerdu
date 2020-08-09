package DLinkList

type Node struct {
	next  *Node
	prev  *Node
	Key   string
	Value string
	Freq  int
}

type DLinkedList struct {
	head *Node
	tail *Node
	Size int
}

func NewLinkedList() *DLinkedList {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head
	return &DLinkedList{
		head: head,
		tail: tail,
	}
}

func (c *DLinkedList) AddNode(node *Node) {
	next := c.head.next
	c.head.next = node
	next.prev = node
	node.next = next
	node.prev = c.head
	c.Size++
}

func (c *DLinkedList) RemoveNode(node *Node) {
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
	c.Size--
}

func (c *DLinkedList) PopTail() *Node {
	prev := c.tail.prev
	c.RemoveNode(prev)
	return prev
}
