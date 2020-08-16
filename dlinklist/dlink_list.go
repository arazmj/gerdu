// This package implement doubly linked list as backing data structure for cache operations
package dlinklist

// Node data structure
type Node struct {
	next  *Node
	prev  *Node
	Key   string
	Value string
	Freq  int
}

// DLinkedList data structure
type DLinkedList struct {
	head *Node
	tail *Node
	Size int
}

// NewLinkedList constructor
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

// AddNode adds a new node to the tail of of linked list
func (c *DLinkedList) AddNode(node *Node) {
	next := c.head.next
	c.head.next = node
	next.prev = node
	node.next = next
	node.prev = c.head
	c.Size++
}

// RemoveNode removes a node
func (c *DLinkedList) RemoveNode(node *Node) {
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
	c.Size--
}

// PopTail pops a node from the beginning of the linked list
func (c *DLinkedList) PopTail() *Node {
	prev := c.tail.prev
	c.RemoveNode(prev)
	return prev
}
