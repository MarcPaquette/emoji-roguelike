package ecs

import "testing"

// stub component used only in tests
type testComp struct{ val int }

func (testComp) Type() ComponentType { return 1 }

type otherComp struct{}

func (otherComp) Type() ComponentType { return 2 }

func TestCreateEntity(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()
	if id == NilEntity {
		t.Fatal("expected non-nil entity ID")
	}
	if !w.Alive(id) {
		t.Fatal("expected entity to be alive after creation")
	}
}

func TestAddAndGetComponent(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()
	w.Add(id, testComp{val: 42})

	c := w.Get(id, ComponentType(1))
	if c == nil {
		t.Fatal("expected component, got nil")
	}
	tc, ok := c.(testComp)
	if !ok {
		t.Fatal("wrong component type returned")
	}
	if tc.val != 42 {
		t.Fatalf("expected val=42, got %d", tc.val)
	}
}

func TestDestroyEntityRemovesComponents(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()
	w.Add(id, testComp{val: 7})
	w.DestroyEntity(id)

	if w.Alive(id) {
		t.Fatal("entity should not be alive after DestroyEntity")
	}
	if w.Get(id, ComponentType(1)) != nil {
		t.Fatal("component should be gone after DestroyEntity")
	}
}

func TestQueryFiltersCorrectly(t *testing.T) {
	w := NewWorld()

	// entity with both A and B
	both := w.CreateEntity()
	w.Add(both, testComp{})
	w.Add(both, otherComp{})

	// entity with only A
	onlyA := w.CreateEntity()
	w.Add(onlyA, testComp{})

	results := w.Query(ComponentType(1), ComponentType(2))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != both {
		t.Fatalf("expected entity %v in results, got %v", both, results[0])
	}
}

func TestRemoveComponent(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()
	w.Add(id, testComp{val: 5})

	w.Remove(id, ComponentType(1))

	if w.Get(id, ComponentType(1)) != nil {
		t.Fatal("component should be nil after Remove")
	}
}

func TestRemoveNonexistentIsNoop(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()
	// Removing a component type that was never added must not panic.
	w.Remove(id, ComponentType(99))
}

func TestHasComponent(t *testing.T) {
	w := NewWorld()
	id := w.CreateEntity()

	if w.Has(id, ComponentType(1)) {
		t.Fatal("Has should return false before Add")
	}
	w.Add(id, testComp{val: 1})
	if !w.Has(id, ComponentType(1)) {
		t.Fatal("Has should return true after Add")
	}
	w.Remove(id, ComponentType(1))
	if w.Has(id, ComponentType(1)) {
		t.Fatal("Has should return false after Remove")
	}
}

func TestQueryExcludesDeadEntities(t *testing.T) {
	w := NewWorld()
	alive := w.CreateEntity()
	w.Add(alive, testComp{})

	dead := w.CreateEntity()
	w.Add(dead, testComp{})
	w.DestroyEntity(dead)

	results := w.Query(ComponentType(1))
	for _, id := range results {
		if id == dead {
			t.Fatal("Query returned a destroyed entity")
		}
	}
	if len(results) != 1 || results[0] != alive {
		t.Fatalf("expected only the alive entity; got %v", results)
	}
}
