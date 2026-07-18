UPDATE academic_grade_scale SET max_score = 80 WHERE letter = 'A-';
UPDATE academic_grade_scale SET max_score = 75 WHERE letter = 'B+';
UPDATE academic_grade_scale SET max_score = 70 WHERE letter = 'B';
UPDATE academic_grade_scale SET max_score = 65 WHERE letter = 'B-';
UPDATE academic_grade_scale SET max_score = 60 WHERE letter = 'C+';
UPDATE academic_grade_scale SET max_score = 55 WHERE letter = 'C';
UPDATE academic_grade_scale SET max_score = 50 WHERE letter = 'D';
UPDATE academic_grade_scale SET max_score = 40 WHERE letter = 'F';

CREATE OR REPLACE FUNCTION seed_academic_grade_scale(target_organization_id TEXT)
RETURNS VOID AS $$
BEGIN
    INSERT INTO academic_grade_scale
        (id, organization_id, letter, min_score, max_score, grade_point, passing, sort_order)
    VALUES
        (gen_random_uuid()::TEXT, target_organization_id, 'A',  80, 100, 4.00, true,  1),
        (gen_random_uuid()::TEXT, target_organization_id, 'A-', 75, 80, 3.67, true,  2),
        (gen_random_uuid()::TEXT, target_organization_id, 'B+', 70, 75, 3.33, true,  3),
        (gen_random_uuid()::TEXT, target_organization_id, 'B',  65, 70, 3.00, true,  4),
        (gen_random_uuid()::TEXT, target_organization_id, 'B-', 60, 65, 2.67, true,  5),
        (gen_random_uuid()::TEXT, target_organization_id, 'C+', 55, 60, 2.33, true,  6),
        (gen_random_uuid()::TEXT, target_organization_id, 'C',  50, 55, 2.00, true,  7),
        (gen_random_uuid()::TEXT, target_organization_id, 'D',  40, 50, 1.00, true,  8),
        (gen_random_uuid()::TEXT, target_organization_id, 'F',   0, 40, 0.00, false, 9)
    ON CONFLICT (organization_id, letter) DO NOTHING;
END;
$$ LANGUAGE plpgsql;
