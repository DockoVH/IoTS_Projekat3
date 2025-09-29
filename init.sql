CREATE TABLE IF NOT EXISTS senzor_podaci (
    id SERIAL PRIMARY KEY,
    vreme TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    temperatura DECIMAL(5,2),
    vlaznost_vazduha DECIMAL(5,2),
    pm2_5 DECIMAL(8,3),
    pm10 DECIMAL(8,3)
);

INSERT INTO senzor_podaci (id, vreme, temperatura, vlaznost_vazduha, pm2_5, pm10) values (0, NOW(), 0, 0, 0, 0) ON CONFLICT DO NOTHING;
