package assets

import "testing"

func TestXPCurveStrictlyIncreasing(t *testing.T) {
	prev := 0
	for level := 2; level <= MaxLevel; level++ {
		xp := XPForLevel(level)
		if xp <= prev {
			t.Errorf("XPForLevel(%d)=%d is not greater than XPForLevel(%d)=%d", level, xp, level-1, prev)
		}
		prev = xp
	}
}

func TestXPForLevelEdgeCases(t *testing.T) {
	if XPForLevel(0) != 0 {
		t.Errorf("XPForLevel(0) = %d; want 0", XPForLevel(0))
	}
	if XPForLevel(1) != 0 {
		t.Errorf("XPForLevel(1) = %d; want 0", XPForLevel(1))
	}
	if XPForLevel(2) != 112 {
		t.Errorf("XPForLevel(2) = %d; want 112", XPForLevel(2))
	}
}

func TestXPForKillScales(t *testing.T) {
	xp1 := XPForKill(3, 1)
	xp5 := XPForKill(3, 5)
	if xp5 <= xp1 {
		t.Errorf("XPForKill(3,5)=%d should be > XPForKill(3,1)=%d", xp5, xp1)
	}
}

func TestXPForFloorEntry(t *testing.T) {
	if XPForFloorEntry(0) != 0 {
		t.Errorf("XPForFloorEntry(0) = %d; want 0", XPForFloorEntry(0))
	}
	if XPForFloorEntry(1) != 150 {
		t.Errorf("XPForFloorEntry(1) = %d; want 150", XPForFloorEntry(1))
	}
}

func TestAllClassesHaveGrowth(t *testing.T) {
	for _, c := range Classes {
		if _, ok := ClassGrowths[c.ID]; !ok {
			t.Errorf("class %q has no entry in ClassGrowths", c.ID)
		}
	}
}

func TestGrowthForLevelZeroAtLevel1(t *testing.T) {
	for _, c := range Classes {
		g := ClassGrowths[c.ID]
		hp, atk, def := GrowthForLevel(g, 1)
		if hp != 0 || atk != 0 || def != 0 {
			t.Errorf("class %q: GrowthForLevel(1) = (%d,%d,%d); want (0,0,0)", c.ID, hp, atk, def)
		}
	}
}

func TestGrowthForLevelIncreases(t *testing.T) {
	for _, c := range Classes {
		g := ClassGrowths[c.ID]
		hp1, _, _ := GrowthForLevel(g, 5)
		hp2, _, _ := GrowthForLevel(g, 10)
		if hp2 <= hp1 {
			t.Errorf("class %q: hp at level 10 (%d) should be > level 5 (%d)", c.ID, hp2, hp1)
		}
	}
}

func TestNoDuplicateSkillIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, s := range AllSkills {
		if seen[s.ID] {
			t.Errorf("duplicate skill ID: %q", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestAllClassesHave5NoviceSkills(t *testing.T) {
	for _, c := range Classes {
		skills := SkillsForClass(c.ID, TierNovice, "")
		if len(skills) != 5 {
			t.Errorf("class %q has %d novice skills; want 5", c.ID, len(skills))
		}
	}
}

func TestAllClassesHave10AdeptSkills(t *testing.T) {
	for _, c := range Classes {
		skillsA := SkillsForClass(c.ID, TierAdept, "A")
		skillsB := SkillsForClass(c.ID, TierAdept, "B")
		total := len(skillsA) + len(skillsB)
		if total != 10 {
			t.Errorf("class %q has %d adept skills (A=%d, B=%d); want 10", c.ID, total, len(skillsA), len(skillsB))
		}
	}
}

func TestAllClassesHaveSpecs(t *testing.T) {
	for _, c := range Classes {
		specs, ok := ClassSpecs[c.ID]
		if !ok {
			t.Errorf("class %q missing from ClassSpecs", c.ID)
			continue
		}
		pair, ok := specs[10]
		if !ok {
			t.Errorf("class %q missing level-10 specialization", c.ID)
			continue
		}
		if pair[0].Branch != "A" || pair[1].Branch != "B" {
			t.Errorf("class %q level-10 specs: branches = (%q,%q); want (A,B)", c.ID, pair[0].Branch, pair[1].Branch)
		}
	}
}

func TestSkillByIDFound(t *testing.T) {
	s := SkillByID("arc_t0_atk")
	if s == nil {
		t.Fatal("SkillByID(arc_t0_atk) returned nil")
	}
	if s.Name != "Arcane Attunement" {
		t.Errorf("Name = %q; want Arcane Attunement", s.Name)
	}
}

func TestSkillByIDNotFound(t *testing.T) {
	if SkillByID("nonexistent") != nil {
		t.Error("SkillByID(nonexistent) should return nil")
	}
}

func TestAvailableSkillsExcludesLearned(t *testing.T) {
	learned := map[string]bool{"arc_t0_atk": true}
	skills := AvailableSkills("arcanist", 5, "", learned)
	for _, s := range skills {
		if s.ID == "arc_t0_atk" {
			t.Error("learned skill arc_t0_atk should not appear in available skills")
		}
	}
}

func TestAvailableSkillsIncludesBranchSkills(t *testing.T) {
	skills := AvailableSkills("arcanist", 15, "A", nil)
	hasBranchA := false
	for _, s := range skills {
		if s.Branch == "A" {
			hasBranchA = true
		}
		if s.Branch == "B" {
			t.Error("branch B skills should not appear when branch=A")
		}
	}
	if !hasBranchA {
		t.Error("should include branch A adept skills at level 15")
	}
}

func TestTierForLevel(t *testing.T) {
	cases := []struct {
		level int
		want  SkillTier
	}{
		{1, TierNovice},
		{9, TierNovice},
		{10, TierAdept},
		{24, TierAdept},
		{25, TierExpert},
		{44, TierExpert},
		{45, TierMaster},
		{69, TierMaster},
		{70, TierLegend},
		{100, TierLegend},
	}
	for _, tc := range cases {
		if got := TierForLevel(tc.level); got != tc.want {
			t.Errorf("TierForLevel(%d) = %d; want %d", tc.level, got, tc.want)
		}
	}
}

func TestIsBranchMilestone(t *testing.T) {
	milestones := map[int]bool{10: true, 25: true, 45: true, 70: true}
	for level := 1; level <= 100; level++ {
		got := IsBranchMilestone(level)
		want := milestones[level]
		if got != want {
			t.Errorf("IsBranchMilestone(%d) = %v; want %v", level, got, want)
		}
	}
}
