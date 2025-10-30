CREATE TABLE IF NOT EXISTS public.outbox (
			  id uuid PRIMARY KEY 
				,body jsonb NOT NULL
);
