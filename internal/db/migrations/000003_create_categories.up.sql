CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    parent_id   INT REFERENCES categories(id) ON DELETE SET NULL,
    description TEXT,
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_categories (
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    confidence  NUMERIC(3,2) DEFAULT 1.0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, category_id)
);

CREATE INDEX idx_project_categories_category ON project_categories(category_id);

-- Seed initial categories
INSERT INTO categories (name, slug, description, sort_order) VALUES
    ('大语言模型', 'llm', 'LLM、ChatBot、文本生成相关项目', 1),
    ('AI Agent', 'agent', 'AI Agent 框架、自主代理、Agentic 工具', 2),
    ('RAG', 'rag', '检索增强生成、知识库、文档问答', 3),
    ('图像生成', 'diffusion', 'Diffusion 模型、文生图、图像编辑', 4),
    ('MLOps', 'mlops', 'ML 工程化、模型训练/部署/监控', 5),
    ('向量数据库', 'vector-db', '向量存储、相似度搜索、Embedding', 6),
    ('AI 框架', 'framework', '深度学习框架、训练工具', 7),
    ('AI 工具', 'tool', 'AI 辅助开发、代码生成、AI 助手', 8),
    ('多模态', 'multimodal', '多模态模型、视觉语言模型', 9),
    ('语音', 'speech', 'TTS、ASR、语音克隆', 10),
    ('强化学习', 'rl', 'RLHF、强化学习框架', 11),
    ('其他', 'other', '未归类的 AI 相关项目', 99);
