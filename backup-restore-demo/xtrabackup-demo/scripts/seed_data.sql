-- ============================================================
-- 初始化演示数据（由 01_full_backup.sh 自动调用）
-- ============================================================
USE shopdb;

CREATE TABLE IF NOT EXISTS orders (
  id         INT AUTO_INCREMENT PRIMARY KEY,
  customer   VARCHAR(100) NOT NULL,
  product    VARCHAR(100) NOT NULL,
  amount     DECIMAL(10,2) NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 插入第一批数据（全量备份前）
INSERT INTO orders (customer, product, amount) VALUES
  ('Alice',   'MacBook Pro',  12999.00),
  ('Bob',     'iPhone 15',     7999.00),
  ('Charlie', 'AirPods Pro',   1799.00),
  ('Diana',   'iPad Air',      4799.00),
  ('Eve',     'Apple Watch',   2999.00);

SELECT '✅ 初始数据写入完成' AS status;
SELECT * FROM orders;
