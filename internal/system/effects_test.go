package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"testing"
)

// newEffectsWorld creates a fresh world with an entity that has the given active effects.
func newEffectsWorld(effects ...component.ActiveEffect) (*ecs.World, ecs.EntityID) {
	w := ecs.NewWorld()
	id := w.CreateEntity()
	w.Add(id, component.Effects{Active: effects})
	return w, id
}

func TestTickEffectsDecrement(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 2, TurnsRemaining: 3,
	})
	TickEffects(w)
	effs := w.Get(id, component.CEffects).(component.Effects)
	if len(effs.Active) != 1 {
		t.Fatalf("expected 1 active effect, got %d", len(effs.Active))
	}
	if effs.Active[0].TurnsRemaining != 2 {
		t.Errorf("TurnsRemaining = %d; want 2", effs.Active[0].TurnsRemaining)
	}
}

func TestTickEffectsExpiry(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 1, TurnsRemaining: 1,
	})
	TickEffects(w)
	effs := w.Get(id, component.CEffects).(component.Effects)
	if len(effs.Active) != 0 {
		t.Errorf("expected 0 active effects after expiry, got %d", len(effs.Active))
	}
}

func TestTickEffectsMultiple(t *testing.T) {
	// Poison expires (1 turn remaining), Weaken survives (3 turns remaining).
	w, id := newEffectsWorld(
		component.ActiveEffect{Kind: component.EffectPoison, Magnitude: 1, TurnsRemaining: 1},
		component.ActiveEffect{Kind: component.EffectWeaken, Magnitude: 2, TurnsRemaining: 3},
	)
	TickEffects(w)
	effs := w.Get(id, component.CEffects).(component.Effects)
	if len(effs.Active) != 1 {
		t.Fatalf("expected 1 effect to remain, got %d", len(effs.Active))
	}
	if effs.Active[0].Kind != component.EffectWeaken {
		t.Errorf("expected EffectWeaken to survive, got %v", effs.Active[0].Kind)
	}
	if effs.Active[0].TurnsRemaining != 2 {
		t.Errorf("TurnsRemaining = %d; want 2", effs.Active[0].TurnsRemaining)
	}
}

func TestApplyEffectCreatesComponent(t *testing.T) {
	w := ecs.NewWorld()
	id := w.CreateEntity()
	// Entity has no CEffects yet — ApplyEffect must create it.
	ApplyEffect(w, id, component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 2, TurnsRemaining: 5,
	})
	c := w.Get(id, component.CEffects)
	if c == nil {
		t.Fatal("expected CEffects component to be created by ApplyEffect")
	}
	effs := c.(component.Effects)
	if len(effs.Active) != 1 || effs.Active[0].Kind != component.EffectPoison {
		t.Errorf("unexpected active effects: %v", effs.Active)
	}
}

func TestApplyEffectReplacesWhenLonger(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 1, TurnsRemaining: 2,
	})
	// Longer duration should replace the existing shorter one.
	ApplyEffect(w, id, component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 3, TurnsRemaining: 5,
	})
	effs := w.Get(id, component.CEffects).(component.Effects)
	if len(effs.Active) != 1 {
		t.Fatalf("expected 1 active effect, got %d", len(effs.Active))
	}
	if effs.Active[0].TurnsRemaining != 5 {
		t.Errorf("TurnsRemaining = %d; want 5 (longer duration should win)", effs.Active[0].TurnsRemaining)
	}
}

func TestApplyEffectSkipsWhenShorter(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 3, TurnsRemaining: 5,
	})
	// Shorter duration must NOT replace the existing longer one.
	ApplyEffect(w, id, component.ActiveEffect{
		Kind: component.EffectPoison, Magnitude: 1, TurnsRemaining: 2,
	})
	effs := w.Get(id, component.CEffects).(component.Effects)
	if effs.Active[0].TurnsRemaining != 5 {
		t.Errorf("TurnsRemaining = %d; shorter effect must not replace the longer one", effs.Active[0].TurnsRemaining)
	}
}

func TestApplyEffectDifferentKindsStack(t *testing.T) {
	w := ecs.NewWorld()
	id := w.CreateEntity()
	ApplyEffect(w, id, component.ActiveEffect{Kind: component.EffectPoison, Magnitude: 1, TurnsRemaining: 3})
	ApplyEffect(w, id, component.ActiveEffect{Kind: component.EffectWeaken, Magnitude: 2, TurnsRemaining: 4})
	effs := w.Get(id, component.CEffects).(component.Effects)
	if len(effs.Active) != 2 {
		t.Errorf("expected 2 distinct effects, got %d", len(effs.Active))
	}
}

func TestHasEffect(t *testing.T) {
	cases := []struct {
		name  string
		setup func() (*ecs.World, ecs.EntityID)
		kind  component.EffectKind
		want  bool
	}{
		{
			name: "entity has the effect",
			setup: func() (*ecs.World, ecs.EntityID) {
				return newEffectsWorld(component.ActiveEffect{Kind: component.EffectPoison, TurnsRemaining: 3})
			},
			kind: component.EffectPoison,
			want: true,
		},
		{
			name: "entity has a different effect",
			setup: func() (*ecs.World, ecs.EntityID) {
				return newEffectsWorld(component.ActiveEffect{Kind: component.EffectWeaken, TurnsRemaining: 3})
			},
			kind: component.EffectPoison,
			want: false,
		},
		{
			name: "entity has no CEffects component",
			setup: func() (*ecs.World, ecs.EntityID) {
				w := ecs.NewWorld()
				return w, w.CreateEntity()
			},
			kind: component.EffectPoison,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, id := tc.setup()
			if got := HasEffect(w, id, tc.kind); got != tc.want {
				t.Errorf("HasEffect = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestGetAttackBonus(t *testing.T) {
	cases := []struct {
		name    string
		effects []component.ActiveEffect
		want    int
	}{
		{
			name:    "no CEffects component",
			effects: nil,
			want:    0,
		},
		{
			name:    "attack boost adds to total",
			effects: []component.ActiveEffect{{Kind: component.EffectAttackBoost, Magnitude: 3, TurnsRemaining: 2}},
			want:    3,
		},
		{
			name:    "weaken subtracts from total",
			effects: []component.ActiveEffect{{Kind: component.EffectWeaken, Magnitude: 2, TurnsRemaining: 2}},
			want:    -2,
		},
		{
			name: "boost and weaken cancel partially",
			effects: []component.ActiveEffect{
				{Kind: component.EffectAttackBoost, Magnitude: 5, TurnsRemaining: 2},
				{Kind: component.EffectWeaken, Magnitude: 2, TurnsRemaining: 2},
			},
			want: 3,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var w *ecs.World
			var id ecs.EntityID
			if tc.effects == nil {
				w = ecs.NewWorld()
				id = w.CreateEntity()
			} else {
				w, id = newEffectsWorld(tc.effects...)
			}
			if got := GetAttackBonus(w, id); got != tc.want {
				t.Errorf("GetAttackBonus = %d; want %d", got, tc.want)
			}
		})
	}
}

func TestGetDefenseBonus(t *testing.T) {
	cases := []struct {
		name    string
		effects []component.ActiveEffect
		want    int
	}{
		{name: "no effects", effects: nil, want: 0},
		{
			name:    "defense boost",
			effects: []component.ActiveEffect{{Kind: component.EffectDefenseBoost, Magnitude: 4, TurnsRemaining: 2}},
			want:    4,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var w *ecs.World
			var id ecs.EntityID
			if tc.effects == nil {
				w = ecs.NewWorld()
				id = w.CreateEntity()
			} else {
				w, id = newEffectsWorld(tc.effects...)
			}
			if got := GetDefenseBonus(w, id); got != tc.want {
				t.Errorf("GetDefenseBonus = %d; want %d", got, tc.want)
			}
		})
	}
}

func TestGetPoisonDamage(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{Kind: component.EffectPoison, Magnitude: 3, TurnsRemaining: 5})
	if got := GetPoisonDamage(w, id); got != 3 {
		t.Errorf("GetPoisonDamage = %d; want 3", got)
	}
	// Entity without CEffects must return 0.
	w2 := ecs.NewWorld()
	id2 := w2.CreateEntity()
	if got := GetPoisonDamage(w2, id2); got != 0 {
		t.Errorf("GetPoisonDamage with no component = %d; want 0", got)
	}
}

func TestGetSelfBurnDamage(t *testing.T) {
	w, id := newEffectsWorld(component.ActiveEffect{Kind: component.EffectSelfBurn, Magnitude: 5, TurnsRemaining: 2})
	if got := GetSelfBurnDamage(w, id); got != 5 {
		t.Errorf("GetSelfBurnDamage = %d; want 5", got)
	}
}

func TestEffectBonusesReturnZeroWithNoComponent(t *testing.T) {
	w := ecs.NewWorld()
	id := w.CreateEntity()
	if v := GetAttackBonus(w, id); v != 0 {
		t.Errorf("GetAttackBonus = %d; want 0 for entity with no CEffects", v)
	}
	if v := GetDefenseBonus(w, id); v != 0 {
		t.Errorf("GetDefenseBonus = %d; want 0 for entity with no CEffects", v)
	}
	if v := GetPoisonDamage(w, id); v != 0 {
		t.Errorf("GetPoisonDamage = %d; want 0 for entity with no CEffects", v)
	}
	if v := GetSelfBurnDamage(w, id); v != 0 {
		t.Errorf("GetSelfBurnDamage = %d; want 0 for entity with no CEffects", v)
	}
}
