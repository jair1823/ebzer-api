ALTER TABLE orders ADD COLUMN platform TEXT NOT NULL DEFAULT 'whatsapp' CHECK(platform IN ('whatsapp', 'instagram', 'facebook'));
