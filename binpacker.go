/*
This file was converted from C++ to Go. The following is the original author's
file comment:

Performs 'discrete online rectangle packing into a rectangular bin' by
maintaining a binary tree of used and free rectangles of the bin. There are
several variants of bin packing problems, and this packer is characterized by:
- We're solving the 'online' version of the problem, which means that when we're
  adding a rectangle, we have no information of the sizes of the rectangles that
  are going to be packed after this one.
- We are packing rectangles that are not rotated. I.e. the algorithm will not
  flip a rectangle of (w,h) to be stored if it were a rectangle of size (h, w).
  There is no restriction conserning UV mapping why this couldn't be done to
  achieve better occupancy, but it's more work. Feel free to try it out.
- The packing is done in discrete integer coordinates and not in rational/real
  numbers (floats).

Internal memory usage is linear to the number of rectangles we've already
packed.

For more information, see
- Rectangle packing: http://www.gamedev.net/community/forums/topic.asp?topic_id=392413
- Packing lightmaps: http://www.blackpawn.com/texts/lightmaps/default.html

Idea: Instead of just picking the first free rectangle to insert the new rect
into, check all free ones (or maintain a sorted order) and pick the one that
minimizes the resulting leftover area. There is no real reason to maintain a
tree - in fact it's just redundant structuring. We could as well have two
lists - one for free rectangles and one for used rectangles. This method would
be faster and might even achieve a considerably better occupancy rate.
*/
package binpacker

import "errors"

func New(width, height int) *Packer {
	return &Packer{
		root:      node{Rect: Rect{Width: width, Height: height}},
		binWidth:  width,
		binHeight: height,
	}
}

type Packer struct {
	root                node
	binWidth, binHeight int
}

type node struct {
	Rect
	left, right *node
}

type Rect struct{ X, Y, Width, Height int }

// Enlarge will mark the previous space as completely occupied and insert the
// new area right and down of the existing area.
func (p *Packer) Enlarge(newWidth, newHeight int) error {
	if newWidth < p.binWidth || newHeight < p.binHeight {
		return errors.New("enlarge: new size is smaller")
	}

	p.root = node{
		Rect: Rect{X: 0, Y: 0, Width: newWidth, Height: newHeight},
		left: &node{Rect: Rect{
			X:      0,
			Y:      p.binHeight,
			Width:  newWidth,
			Height: newHeight - p.binHeight,
		}},
		right: &node{Rect: Rect{
			X:      p.binWidth,
			Y:      0,
			Width:  newWidth - p.binWidth,
			Height: p.binHeight,
		}},
	}

	p.binWidth = newWidth
	p.binHeight = newHeight

	return nil
}

func (p *Packer) Insert(width, height int) (Rect, error) {
	n, err := insert(&p.root, width, height)
	if err != nil {
		return Rect{}, err
	}
	return n.Rect, nil
}

var ErrNoMoreSpace = errors.New("insert: no more space in bin")

func insert(n *node, width, height int) (*node, error) {
	if n.left != nil || n.right != nil {
		if n.left != nil {
			newNode, _ := insert(n.left, width, height)
			if newNode != nil {
				return newNode, nil
			}
		}
		if n.right != nil {
			newNode, _ := insert(n.right, width, height)
			if newNode != nil {
				return newNode, nil
			}
		}
		return nil, ErrNoMoreSpace
	}

	// this node is a leaf, can we insert the new rectangle here?
	if width > n.Width || height > n.Height {
		return nil, ErrNoMoreSpace
	}

	// the new cell will fit, split the remaining space along the shorter axis,
	// that is probably more optimal.
	restW, restH := n.Width-width, n.Height-height

	if restW < restH {
		// split the remaining space horizontally
		n.left = &node{Rect: Rect{
			X:      n.X + width,
			Y:      n.Y,
			Width:  restW,
			Height: height,
		}}
		n.right = &node{Rect: Rect{
			X:      n.X,
			Y:      n.Y + height,
			Width:  n.Width,
			Height: restH,
		}}
	} else {
		// split the remaining space vertically
		n.left = &node{Rect: Rect{
			X:      n.X,
			Y:      n.Y + height,
			Width:  width,
			Height: restH,
		}}
		n.right = &node{Rect: Rect{
			X:      n.X + width,
			Y:      n.Y,
			Width:  restW,
			Height: n.Height,
		}}
	}

	// Note that as a result of the above, it can happen that node->left or
	// node->right is now a degenerate (zero area) rectangle. No need to do
	// anything about it, like remove the nodes as "unnecessary" since they
	// need to exist as children of this node (this node can't be a leaf
	// anymore).

	// This node is now a non-leaf, so shrink its area - it now denotes
	// *occupied* space instead of free space. Its children spawn the resulting
	// area of free space.
	n.Width, n.Height = width, height
	return n, nil
}

func (p *Packer) Occupancy() float64 {
	return float64(usedArea(&p.root)) / float64(p.binWidth*p.binHeight)
}

func usedArea(n *node) int {
	if n.left != nil || n.right != nil {
		used := n.Width * n.Height
		if n.left != nil {
			used += usedArea(n.left)
		}
		if n.right != nil {
			used += usedArea(n.right)
		}
		return used
	}
	// this is a leaf node and does not constitute to the total surface area
	return 0
}
