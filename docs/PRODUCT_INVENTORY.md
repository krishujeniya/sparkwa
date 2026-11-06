# SparkWA documentation content inventory

This is a **single-product repository**. There is no multi-product registry, pricing tier, or product explorer.

| Public content | Input source | Generated route | Category | Publication |
|---|---|---|---|---|
| SparkWA home and product explanation | `README.md`, `docs/index.html` | `/` | Overview | Public |
| Documentation overview | Explicit guide allowlist | `/docs/` | Reference | Public |
| Install on a VPS | `docs/VPS_INSTALL.md` | `/docs/vps-install/` | Get started | Public |
| Connect Gemini Spark | `docs/SPARK_SETUP.md` | `/docs/spark-setup/` | Connect | Public |
| Gmail OAuth | `docs/GMAIL_OAUTH_SETUP.md` | `/docs/gmail-oauth/` | Connect | Public |
| Security | `docs/SECURITY.md` | `/docs/security/` | Operate | Public |
| Troubleshooting | `docs/TROUBLESHOOTING.md` | `/docs/troubleshooting/` | Operate | Public |
| Original full reference | Previously published `docs/index.html` | `/reference/complete-guide.html` | Reference | Public |

## Ingestion policy
- Only the five explicitly listed public Markdown guides are compiled into HTML. **No files under private data, environment values, secrets, local state, tokens, or config are indexed**.
- Routes use fixed stable slugs, so duplicate filenames cannot collide.
- Links to the other five guides are rewritten to the local HTML routes. Original links and text remain available on GitHub.
- Search is a static `docs/assets/search-index.js` generated from the five public docs, the reference, and the overview. It exposes no user data.
- Markdown renders via a deliberately limited deterministic parser for headings, paragraphs, lists, code fences, tables and simple inline syntax; complex Markdown constructs may need manual parity checks.
- `docs/assets/sparkwa-banner.png` and `docs/assets/sparkwa-mark.svg` are project-owned visual assets.
- Original technical sections remain available in the preserved complete reference if concise new overview omits detail.
