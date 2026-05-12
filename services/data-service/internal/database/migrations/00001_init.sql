-- +goose Up
-- +goose StatementBegin

-- 扩展
-- pgcrypto: gen_random_uuid()
-- pg_trgm:  LOWER(col) LIKE '%kw%' 的 GIN 表达式索引；缺失则搜索退化为全表扫描
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 院校
CREATE TABLE IF NOT EXISTS universities (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code               varchar(20)  NOT NULL,
    name               varchar(200) NOT NULL,
    english_name       varchar(300),
    alias              varchar(500),
    type               varchar(50),
    level              varchar(50),
    nature             varchar(50),
    category           varchar(100),
    province           varchar(50),
    city               varchar(50),
    district           varchar(50),
    address            varchar(500),
    postal_code        varchar(20),
    website            varchar(255),
    phone              varchar(50),
    email              varchar(100),
    established        timestamp with time zone,
    description        text,
    motto              varchar(500),
    logo               varchar(255),
    campus_area        double precision,
    student_count      bigint DEFAULT 0,
    teacher_count      bigint DEFAULT 0,
    academician_count  bigint DEFAULT 0,
    national_rank      bigint,
    province_rank      bigint,
    qs_rank            bigint,
    us_news_rank       bigint,
    overall_score      double precision,
    teaching_score     double precision,
    research_score     double precision,
    employment_score   double precision,
    status             varchar(20)  DEFAULT 'active',
    is_active          boolean      DEFAULT true,
    is_recruiting      boolean      DEFAULT true,
    created_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at         timestamp with time zone
);

-- 专业
CREATE TABLE IF NOT EXISTS majors (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    university_id      uuid NOT NULL,
    code               varchar(20),
    name               varchar(200) NOT NULL,
    english_name       varchar(300),
    category           varchar(100),
    discipline         varchar(100),
    sub_discipline     varchar(100),
    degree_type        varchar(50),
    duration           bigint,
    description        text,
    core_courses       text,
    requirements       text,
    career_prospects   text,
    employment_rate    double precision,
    average_salary     double precision,
    top_employers      text,
    is_recruiting      boolean DEFAULT true,
    recruitment_quota  bigint,
    view_count         bigint DEFAULT 0,
    search_count       bigint DEFAULT 0,
    popularity_score   double precision DEFAULT 0,
    status             varchar(20) DEFAULT 'active',
    is_active          boolean     DEFAULT true,
    created_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at         timestamp with time zone
);

-- 录取数据
CREATE TABLE IF NOT EXISTS admission_data (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    university_id   uuid NOT NULL,
    major_id        uuid,
    year            bigint NOT NULL,
    province        varchar(50),
    batch           varchar(50),
    category        varchar(50),
    min_score       double precision,
    max_score       double precision,
    avg_score       double precision,
    median_score    double precision,
    min_rank        bigint,
    max_rank        bigint,
    avg_rank        bigint,
    planned_count   bigint,
    actual_count    bigint,
    difficulty      varchar(20),
    competition     double precision,
    admission_rate  double precision,
    created_at      timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at      timestamp with time zone
);

-- 搜索索引
CREATE TABLE IF NOT EXISTS search_indices (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type            varchar(50),
    entity_id       uuid NOT NULL,
    title           varchar(500) NOT NULL,
    content         text,
    keywords        text,
    tags            text,
    province        varchar(50),
    category        varchar(100),
    search_weight   double precision DEFAULT 1.0,
    view_count      bigint DEFAULT 0,
    search_count    bigint DEFAULT 0,
    last_viewed     timestamp with time zone,
    created_at      timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at      timestamp with time zone
);

-- 分析结果
CREATE TABLE IF NOT EXISTS analysis_results (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid,
    request_id    varchar(100),
    province      varchar(50),
    score         double precision,
    rank          bigint,
    category      varchar(50),
    preferences   text,
    results       text,
    confidence    double precision,
    process_time  double precision,
    algorithm     varchar(50),
    version       varchar(20),
    created_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at    timestamp with time zone
);

-- 热门搜索
CREATE TABLE IF NOT EXISTS hot_searches (
    id            bigserial PRIMARY KEY,
    keyword       varchar(200),
    search_count  bigint DEFAULT 0,
    category      varchar(50),
    date          timestamp with time zone,
    created_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 数据统计
CREATE TABLE IF NOT EXISTS data_statistics (
    id            bigserial PRIMARY KEY,
    stat_type     varchar(50),
    stat_key      varchar(200),
    stat_value    double precision,
    string_value  varchar(500),
    json_value    text,
    date          timestamp with time zone,
    created_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 院校统计
CREATE TABLE IF NOT EXISTS university_statistics (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    university_id      uuid NOT NULL,
    national_rank      bigint,
    province_rank      bigint,
    qs_rank            bigint,
    us_news_rank       bigint,
    employment_rate    double precision,
    average_salary     double precision,
    top_employers      text,
    teaching_quality   double precision,
    research_quality   double precision,
    updated_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 专业统计
CREATE TABLE IF NOT EXISTS major_statistics (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    major_id        uuid NOT NULL,
    average_score   double precision,
    min_score       double precision,
    max_score       double precision,
    employment_rate double precision,
    average_salary  double precision,
    popularity      bigint,
    updated_at      timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 单列索引（GORM `gorm:"index"` tag 列）
CREATE INDEX IF NOT EXISTS idx_universities_code              ON universities(code);
CREATE INDEX IF NOT EXISTS idx_universities_name              ON universities(name);
CREATE INDEX IF NOT EXISTS idx_universities_deleted_at        ON universities(deleted_at);
CREATE INDEX IF NOT EXISTS idx_majors_university_id           ON majors(university_id);
CREATE INDEX IF NOT EXISTS idx_majors_code                    ON majors(code);
CREATE INDEX IF NOT EXISTS idx_majors_name                    ON majors(name);
CREATE INDEX IF NOT EXISTS idx_majors_deleted_at              ON majors(deleted_at);
CREATE INDEX IF NOT EXISTS idx_admission_data_university_id   ON admission_data(university_id);
CREATE INDEX IF NOT EXISTS idx_admission_data_major_id        ON admission_data(major_id);
CREATE INDEX IF NOT EXISTS idx_admission_data_deleted_at      ON admission_data(deleted_at);
CREATE INDEX IF NOT EXISTS idx_search_indices_entity_id       ON search_indices(entity_id);
CREATE INDEX IF NOT EXISTS idx_search_indices_title           ON search_indices(title);
CREATE INDEX IF NOT EXISTS idx_search_indices_deleted_at      ON search_indices(deleted_at);
CREATE INDEX IF NOT EXISTS idx_analysis_results_user_id       ON analysis_results(user_id);
CREATE INDEX IF NOT EXISTS idx_analysis_results_request_id    ON analysis_results(request_id);
CREATE INDEX IF NOT EXISTS idx_analysis_results_province      ON analysis_results(province);
CREATE INDEX IF NOT EXISTS idx_analysis_results_deleted_at    ON analysis_results(deleted_at);
CREATE INDEX IF NOT EXISTS idx_hot_searches_keyword           ON hot_searches(keyword);
CREATE INDEX IF NOT EXISTS idx_hot_searches_date              ON hot_searches(date);
CREATE INDEX IF NOT EXISTS idx_data_statistics_stat_type      ON data_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_data_statistics_date           ON data_statistics(date);
CREATE INDEX IF NOT EXISTS idx_university_statistics_uni_id   ON university_statistics(university_id);
CREATE INDEX IF NOT EXISTS idx_major_statistics_major_id      ON major_statistics(major_id);

-- 复合索引（原 createIndices() 中定义）
-- BUGFIX: 旧 idx_universities_rank_score 使用 universities.popularity_score 但该列不存在，
--         CREATE INDEX 在每次启动时静默失败。这里修正为 (national_rank, overall_score)。
-- BUGFIX: 旧 idx_admission_data_year_batch 使用 batch_type 列也不存在，同上静默失败。
--         修正为 (year, batch)。
CREATE INDEX IF NOT EXISTS idx_universities_province_type     ON universities(province, type);
CREATE INDEX IF NOT EXISTS idx_universities_level_nature      ON universities(level, nature);
CREATE INDEX IF NOT EXISTS idx_universities_province_level    ON universities(province, level);
CREATE INDEX IF NOT EXISTS idx_universities_type_level        ON universities(type, level);
CREATE INDEX IF NOT EXISTS idx_universities_active_recruiting ON universities(is_active, is_recruiting);
CREATE INDEX IF NOT EXISTS idx_universities_rank_score        ON universities(national_rank, overall_score);
CREATE INDEX IF NOT EXISTS idx_majors_university_category     ON majors(university_id, category);
CREATE INDEX IF NOT EXISTS idx_majors_discipline_degree       ON majors(discipline, degree_type);
CREATE INDEX IF NOT EXISTS idx_majors_category_active         ON majors(category, is_active);
CREATE INDEX IF NOT EXISTS idx_majors_university_active       ON majors(university_id, is_active);
CREATE INDEX IF NOT EXISTS idx_admission_data_year_province   ON admission_data(year, province);
CREATE INDEX IF NOT EXISTS idx_admission_data_university_year ON admission_data(university_id, year);
CREATE INDEX IF NOT EXISTS idx_admission_data_score_rank      ON admission_data(avg_score, min_rank);
CREATE INDEX IF NOT EXISTS idx_admission_data_year_batch      ON admission_data(year, batch);
CREATE INDEX IF NOT EXISTS idx_search_indices_type_province   ON search_indices(type, province);
CREATE INDEX IF NOT EXISTS idx_analysis_results_user_created  ON analysis_results(user_id, created_at);

-- pg_trgm GIN 表达式索引：覆盖 services 层 LOWER(col) LIKE '%kw%'。
-- 必须建在 LOWER(col) 上 planner 才会用——直接建在 col 上 WHERE 走不到。
CREATE INDEX IF NOT EXISTS idx_universities_name_trgm    ON universities  USING gin (LOWER(name)    gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_universities_code_trgm    ON universities  USING gin (LOWER(code)    gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_universities_alias_trgm   ON universities  USING gin (LOWER(alias)   gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_majors_name_trgm          ON majors        USING gin (LOWER(name)    gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_hot_searches_keyword_trgm ON hot_searches  USING gin (LOWER(keyword) gin_trgm_ops);

-- popularity_score seed：保持与旧 migrate() UPDATE 行为一致。
-- 注：增加 `AND popularity_score = 0` 限定避免覆盖人工设定值；旧版重启时无条件覆盖，是 bug。
UPDATE majors SET popularity_score = 95 WHERE (name LIKE '%计算机%' OR name LIKE '%软件%'   OR name LIKE '%人工智能%') AND popularity_score = 0;
UPDATE majors SET popularity_score = 90 WHERE (name LIKE '%电子%'   OR name LIKE '%通信%'   OR name LIKE '%自动化%')   AND popularity_score = 0;
UPDATE majors SET popularity_score = 85 WHERE (name LIKE '%金融%'   OR name LIKE '%经济%'   OR name LIKE '%管理%')     AND popularity_score = 0;
UPDATE majors SET popularity_score = 80 WHERE (name LIKE '%医学%'   OR name LIKE '%临床%'   OR name LIKE '%护理%')     AND popularity_score = 0;
UPDATE majors SET popularity_score = 75 WHERE (name LIKE '%机械%'   OR name LIKE '%土木%'   OR name LIKE '%建筑%')     AND popularity_score = 0;
UPDATE majors SET popularity_score = 70 WHERE popularity_score = 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS major_statistics      CASCADE;
DROP TABLE IF EXISTS university_statistics CASCADE;
DROP TABLE IF EXISTS data_statistics       CASCADE;
DROP TABLE IF EXISTS hot_searches          CASCADE;
DROP TABLE IF EXISTS analysis_results      CASCADE;
DROP TABLE IF EXISTS search_indices        CASCADE;
DROP TABLE IF EXISTS admission_data        CASCADE;
DROP TABLE IF EXISTS majors                CASCADE;
DROP TABLE IF EXISTS universities          CASCADE;

-- +goose StatementEnd
