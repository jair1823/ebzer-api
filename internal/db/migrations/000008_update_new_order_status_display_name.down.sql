UPDATE order_statuses
SET display_name = 'Nuevo',
    updated_at = datetime('now')
WHERE name = 'new';
