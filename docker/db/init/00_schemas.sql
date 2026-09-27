-- Ejecutado por el entrypoint de Postgres solo en el primer arranque
-- (volumen vacío). El search_path de la app apunta a dw_lubrisur.
CREATE SCHEMA IF NOT EXISTS dw_lubrisur;
