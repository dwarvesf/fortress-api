INSERT INTO public.configs
    (id, deleted_at, created_at, updated_at, key, value)
SELECT '5a02645d-bbfd-4da5-9582-3d4a96bcfeb8', NULL, '2023-12-11 09:50:25.714604', '2023-12-11 09:50:25.714604', 'salary-advance-max-cap', '25'
WHERE NOT EXISTS (
    SELECT 1
FROM public.configs
WHERE key = 'salary-advance-max-cap'
);

INSERT INTO public.configs
    (id, deleted_at, created_at, updated_at, key, value)
SELECT '1899da85-b29a-486c-b54b-8822a085c02f', NULL, '2023-12-11 09:50:25.714604', '2023-12-11 09:50:25.714604', 'icy-usd-rate', '1.5'
WHERE NOT EXISTS (
    SELECT 1
FROM public.configs
WHERE key = 'icy-usd-rate'
);

