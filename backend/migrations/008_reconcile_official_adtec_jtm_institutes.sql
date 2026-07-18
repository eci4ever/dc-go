-- Reconcile the initial ILJTM seed with the official ADTEC JTM campus names.
-- Keep the original ADTEC Melaka record as the Proton institute so any
-- organization-scoped references remain attached to the canonical record.
DELETE FROM organization
WHERE slug = 'proton-institute-adtec-melaka';

WITH official(old_slug, official_name, official_slug) AS (
    VALUES
        ('jmti', 'Institut Teknikal Jepun Malaysia (JMTI)', 'institut-teknikal-jepun-malaysia-jmti'),
        ('adtec-shah-alam', 'ADTEC JTM Kampus Shah Alam', 'adtec-jtm-kampus-shah-alam'),
        ('adtec-batu-pahat', 'ADTEC JTM Kampus Batu Pahat', 'adtec-jtm-kampus-batu-pahat'),
        ('adtec-kulim', 'ADTEC JTM Kampus Kulim', 'adtec-jtm-kampus-kulim'),
        ('adtec-melaka', 'Institut Teknologi Automotif Termaju Proton (ADTEC Melaka)', 'institut-teknologi-automotif-termaju-proton-adtec-melaka'),
        ('adtec-kemaman', 'ADTEC JTM Kampus Kemaman', 'adtec-jtm-kampus-kemaman'),
        ('adtec-taiping', 'ADTEC JTM Kampus Taiping', 'adtec-jtm-kampus-taiping'),
        ('adtec-bintulu', 'ADTEC JTM Kampus Bintulu', 'adtec-jtm-kampus-bintulu'),
        ('adtec-jerantut', 'ADTEC JTM Kampus Jerantut', 'adtec-jtm-kampus-jerantut'),
        ('ilp-kuala-lumpur', 'ADTEC JTM Kampus Kuala Lumpur', 'adtec-jtm-kampus-kuala-lumpur'),
        ('ilp-kota-bharu', 'ADTEC JTM Kampus Kota Bharu', 'adtec-jtm-kampus-kota-bharu'),
        ('ilp-kuala-terengganu', 'ADTEC JTM Kampus Kuala Terengganu', 'adtec-jtm-kampus-kuala-terengganu'),
        ('ilp-kuantan', 'ADTEC JTM Kampus Kuantan', 'adtec-jtm-kampus-kuantan'),
        ('ilp-jitra', 'ADTEC JTM Kampus Jitra', 'adtec-jtm-kampus-jitra'),
        ('ilp-ipoh', 'ADTEC JTM Kampus Ipoh', 'adtec-jtm-kampus-ipoh'),
        ('ilp-bukit-katil', 'ADTEC JTM Kampus Bukit Katil', 'adtec-jtm-kampus-bukit-katil'),
        ('ilp-pasir-gudang', 'ADTEC JTM Kampus Pasir Gudang', 'adtec-jtm-kampus-pasir-gudang'),
        ('ilp-kangar', 'ADTEC JTM Kampus Kangar', 'adtec-jtm-kampus-kangar'),
        ('ilp-pedas', 'ADTEC JTM Kampus Pedas', 'adtec-jtm-kampus-pedas'),
        ('ilp-tangkak', 'ADTEC JTM Kampus Tangkak', 'adtec-jtm-kampus-tangkak'),
        ('ilp-labuan', 'ADTEC JTM Kampus Labuan', 'adtec-jtm-kampus-labuan'),
        ('ilp-kota-kinabalu', 'ADTEC JTM Kampus Kota Kinabalu', 'adtec-jtm-kampus-kota-kinabalu'),
        ('ilp-kota-samarahan', 'ADTEC JTM Kampus Kota Samarahan', 'adtec-jtm-kampus-kota-samarahan'),
        ('ilp-kepala-batas', 'ADTEC JTM Kampus Kepala Batas', 'adtec-jtm-kampus-kepala-batas'),
        ('ilp-kuala-langat', 'ADTEC JTM Kampus Kuala Langat', 'adtec-jtm-kampus-kuala-langat'),
        ('ilp-selandar', 'ADTEC JTM Kampus Selandar', 'adtec-jtm-kampus-selandar'),
        ('ilp-mersing', 'ADTEC JTM Kampus Mersing', 'adtec-jtm-kampus-mersing'),
        ('ilp-marang', 'ADTEC JTM Kampus Marang', 'adtec-jtm-kampus-marang'),
        ('ilp-miri', 'ADTEC JTM Kampus Miri', 'adtec-jtm-kampus-miri'),
        ('ilp-sandakan', 'ADTEC JTM Kampus Sandakan', 'adtec-jtm-kampus-sandakan'),
        ('ilp-perai', 'ADTEC JTM Kampus Perai', 'adtec-jtm-kampus-perai'),
        ('ilp-arumugam-pillai-nibong-tebal', 'ADTEC JTM Kampus Nibong Tebal', 'adtec-jtm-kampus-nibong-tebal')
)
UPDATE organization AS organization
SET name = official.official_name,
    slug = official.official_slug
FROM official
WHERE organization.slug = official.old_slug;

INSERT INTO organization (id, name, slug)
VALUES
    (gen_random_uuid()::TEXT, 'ADTEC JTM Kampus Serian', 'adtec-jtm-kampus-serian'),
    (gen_random_uuid()::TEXT, 'ADTEC JTM Kampus Senai', 'adtec-jtm-kampus-senai')
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;
