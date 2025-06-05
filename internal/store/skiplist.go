package store

import (
	"math/rand"
)

const (
	maxLevel    = 32   // Maximum level for skip list
	probability = 0.25 // Probability for level promotion
)

// skiplistNode represents a node in the skip list
type skiplistNode struct {
	member   string
	score    float64
	forward  []*skiplistNode // Array of forward pointers
	backward *skiplistNode   // Backward pointer for reverse iteration
	level    int             // Current node level
}

// skiplist represents a skip list data structure
type skiplist struct {
	head   *skiplistNode // Header node
	tail   *skiplistNode // Tail node
	length int           // Number of nodes in the skip list
	level  int           // Current maximum level of the skip list
}

// newSkiplist creates a new skip list
func newSkiplist() *skiplist {
	header := &skiplistNode{
		forward: make([]*skiplistNode, maxLevel),
		level:   maxLevel,
	}
	return &skiplist{
		head:  header,
		level: 1,
	}
}

// randomLevel returns a random level for a new node
func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < probability {
		level++
	}
	return level
}

// insert adds or updates a member in the skip list
func (sl *skiplist) insert(score float64, member string) bool {
	update := make([]*skiplistNode, maxLevel) // Update vector
	current := sl.head

	// Find position to insert
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil &&
			(current.forward[i].score < score ||
				(current.forward[i].score == score && current.forward[i].member < member)) {
			current = current.forward[i]
		}
		update[i] = current
	}

	// Get next node at level 0
	current = current.forward[0]

	// If node exists with same member, update score
	if current != nil && current.member == member {
		oldScore := current.score
		current.score = score

		// If score hasn't changed, no need to reposition
		if oldScore == score {
			return false
		}

		// Remove and reinsert if score changed
		sl.delete(oldScore, member)
		return sl.insert(score, member)
	}

	// Insert new node
	level := randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	// Create new node
	newNode := &skiplistNode{
		member:  member,
		score:   score,
		forward: make([]*skiplistNode, level),
		level:   level,
	}

	// Update forward pointers
	for i := 0; i < level; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	// Update backward pointer
	if update[0] == sl.head {
		newNode.backward = nil
	} else {
		newNode.backward = update[0]
	}

	if newNode.forward[0] != nil {
		newNode.forward[0].backward = newNode
	} else {
		sl.tail = newNode
	}

	sl.length++
	return true
}

// delete removes a member from the skip list
func (sl *skiplist) delete(score float64, member string) bool {
	update := make([]*skiplistNode, maxLevel)
	current := sl.head

	// Find node to delete
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil &&
			(current.forward[i].score < score ||
				(current.forward[i].score == score && current.forward[i].member < member)) {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	// If node doesn't exist or doesn't match
	if current == nil || current.member != member {
		return false
	}

	// Update forward pointers
	for i := 0; i < sl.level; i++ {
		if update[i].forward[i] != current {
			break
		}
		update[i].forward[i] = current.forward[i]
	}

	// Update backward pointer of next node
	if current.forward[0] != nil {
		current.forward[0].backward = current.backward
	} else {
		sl.tail = current.backward
	}

	// Update skip list level
	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}

	sl.length--
	return true
}

// getRange returns a slice of skiplistNodes from start to stop (inclusive).
// If the range exceeds the number of elements in the skiplist, it returns
// as many elements as are available from the start index onward.
func (sl *skiplist) getRange(start, stop int) []*skiplistNode {
	var result []*skiplistNode

	// Handle negative indices
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}

	// Boundary checks
	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}
	if start > stop {
		return result
	}

	// Find start node
	current := sl.head.forward[0]
	for i := 0; i < start && current != nil; i++ {
		current = current.forward[0]
	}

	// Collect nodes
	for i := start; i <= stop && current != nil; i++ {
		result = append(result, current)
		current = current.forward[0]
	}

	return result
}
