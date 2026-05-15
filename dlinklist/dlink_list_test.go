package dlinklist

import "testing"

func assertList(t *testing.T, list *DLinkedList, want []string) {
	t.Helper()

	if got := list.Size(); got != len(want) {
		t.Fatalf("Size() = %d, want %d", got, len(want))
	}

	var gotForward []string
	for node := list.head.next; node != list.tail; node = node.next {
		if node == nil {
			t.Fatalf("forward traversal hit nil before tail; got %v, want %v", gotForward, want)
		}
		gotForward = append(gotForward, node.Key)
	}
	if len(gotForward) != len(want) {
		t.Fatalf("forward traversal length = %d (%v), want %d (%v)", len(gotForward), gotForward, len(want), want)
	}
	for i := range want {
		if gotForward[i] != want[i] {
			t.Fatalf("forward traversal = %v, want %v", gotForward, want)
		}
	}

	for i, node := len(want)-1, list.tail.prev; node != list.head; i, node = i-1, node.prev {
		if node == nil {
			t.Fatalf("backward traversal hit nil before head")
		}
		if i < 0 {
			t.Fatalf("backward traversal has extra node %q", node.Key)
		}
		if node.Key != want[i] {
			t.Fatalf("backward traversal key = %q at reverse index %d, want %q", node.Key, i, want[i])
		}
	}
}

func newNode(key string) *Node {
	return &Node{Key: key, Value: "value-" + key, Freq: len(key)}
}

func TestNewLinkedListIsEmpty(t *testing.T) {
	list := NewLinkedList()

	assertList(t, list, nil)
	if list.head == nil || list.tail == nil {
		t.Fatal("NewLinkedList() did not initialize sentinel nodes")
	}
	if list.head.next != list.tail {
		t.Fatal("empty list head should point to tail")
	}
	if list.tail.prev != list.head {
		t.Fatal("empty list tail should point to head")
	}
}

func TestAddNodeAddsToFrontAndTracksSize(t *testing.T) {
	list := NewLinkedList()
	first := newNode("first")
	second := newNode("second")
	third := newNode("third")

	list.AddNode(first)
	assertList(t, list, []string{"first"})
	if first.prev != list.head || first.next != list.tail {
		t.Fatal("single added node should be linked between head and tail")
	}

	list.AddNode(second)
	list.AddNode(third)
	assertList(t, list, []string{"third", "second", "first"})
}

func TestRemoveNodeRemovesHeadTailMiddleAndOnlyElement(t *testing.T) {
	t.Run("head", func(t *testing.T) {
		list := NewLinkedList()
		a, b, c := newNode("a"), newNode("b"), newNode("c")
		list.AddNode(a)
		list.AddNode(b)
		list.AddNode(c)

		list.RemoveNode(c)
		assertList(t, list, []string{"b", "a"})
	})

	t.Run("tail", func(t *testing.T) {
		list := NewLinkedList()
		a, b, c := newNode("a"), newNode("b"), newNode("c")
		list.AddNode(a)
		list.AddNode(b)
		list.AddNode(c)

		list.RemoveNode(a)
		assertList(t, list, []string{"c", "b"})
	})

	t.Run("middle", func(t *testing.T) {
		list := NewLinkedList()
		a, b, c := newNode("a"), newNode("b"), newNode("c")
		list.AddNode(a)
		list.AddNode(b)
		list.AddNode(c)

		list.RemoveNode(b)
		assertList(t, list, []string{"c", "a"})
	})

	t.Run("only element", func(t *testing.T) {
		list := NewLinkedList()
		only := newNode("only")
		list.AddNode(only)

		list.RemoveNode(only)
		assertList(t, list, nil)
		if list.head.next != list.tail || list.tail.prev != list.head {
			t.Fatal("removing the only element should restore empty sentinel links")
		}
	})
}

func TestPopTailRemovesLeastRecentlyAddedNode(t *testing.T) {
	list := NewLinkedList()
	oldest := newNode("oldest")
	middle := newNode("middle")
	newest := newNode("newest")
	list.AddNode(oldest)
	list.AddNode(middle)
	list.AddNode(newest)

	if got := list.PopTail(); got != oldest {
		t.Fatalf("PopTail() = %q, want oldest node", got.Key)
	}
	assertList(t, list, []string{"newest", "middle"})

	if got := list.PopTail(); got != middle {
		t.Fatalf("PopTail() = %q, want middle node", got.Key)
	}
	assertList(t, list, []string{"newest"})
}
