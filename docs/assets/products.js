// @ts-check
/**
 * The single source of truth for the ecosystem.
 * Product data is grounded in public project README files and repository assets.
 * No inferred billing, deployment, metrics, or customer claims.
 *
 * @typedef {"Agents"|"Developer tools"|"Applied AI"|"Applications"} Category
 * @typedef {"Source available"|"Website referenced"|"Research project"} Availability
 * @typedef {{slug:string,name:string,eyebrow:string,category:Category,tagline:string,description:string,features:readonly string[],audience:string,stack:string,availability:Availability,pricing:string,repo:string,website?:string,docs?:string,media?:string,mediaAlt?:string,color:string,proof:string}} Product
 */
/** @type {readonly Product[]} */
export const products = Object.freeze([
  {
    slug:"hadron", name:"Hadron", eyebrow:"01 / CODE INTELLIGENCE",category:"Developer tools",
    tagline:"A second set of eyes. Four different perspectives.",
    description:"A multi-agent pull request review engine coordinating security, code quality, testing, and documentation specialists, with a human-review queue for lower-confidence findings.",
    features:["Parallel specialist PR review","Repository-aware context retrieval","Confidence scoring and human triage","Next.js review dashboard"],
    audience:"Engineering teams and maintainers",stack:"Python · FastAPI · LangGraph · Next.js",availability:"Website referenced",pricing:"No public subscription price verified",
    repo:"https://github.com/krishujeniya/Hadron",website:"https://hadron.vercel.app/",docs:"https://github.com/krishujeniya/Hadron#readme",
    media:"https://raw.githubusercontent.com/krishujeniya/Hadron/main/docs/visual_audit/transformed-desktop-1280.png",mediaAlt:"Hadron dashboard screenshot from its repository",color:"#95b5ff",proof:"README.md; docs/visual_audit/transformed-desktop-1280.png"
  },
  {
    slug:"agyos",name:"AGYOS",eyebrow:"02 / AGENT OPERATING SYSTEM",category:"Agents",
    tagline:"Give agents a system for getting things done.",
    description:"An agent operating system and catalog of modular engineering runbooks, designed around natural-language orchestration, verification gates, skills, and execution pods.",
    features:["Natural-language task intake","Runbook and skill catalog","Verification and governance gates","Multi-agent operating architecture"],
    audience:"AI agent builders and engineering teams",stack:"Agent workflows · MCP · Automation",availability:"Source available",pricing:"No published SaaS price",
    repo:"https://github.com/krishujeniya/AGYOS",docs:"https://krishujeniya.github.io/AGYOS/",
    media:"https://raw.githubusercontent.com/krishujeniya/AGYOS/main/assets/agyos-logo.svg",mediaAlt:"AGYOS project logo",color:"#b9a2ff",proof:"README.md; assets/agyos-logo.svg"
  },
  {
    slug:"amberclaw",name:"AmberClaw",eyebrow:"03 / PRIVATE AGENT RUNTIME",category:"Agents",
    tagline:"Agent automation with boundaries built in.",
    description:"A self-hosted AI agent runtime with model routing, isolated execution, audit trails, and a web dashboard. Repository documentation describes security and deployment capabilities.",
    features:["Self-hosted agent gateway","Multi-model routing","Audit and approval workflows","Local and container-based execution"],
    audience:"Teams operating private AI automations",stack:"Python · FastAPI · LiteLLM · SQLite",availability:"Source available",pricing:"Self-hosted; cloud pricing not published",
    repo:"https://github.com/krishujeniya/AmberClaw",media:"https://raw.githubusercontent.com/krishujeniya/AmberClaw/main/assets/demo.png",mediaAlt:"AmberClaw demonstration image from its repository",color:"#ffb37c",proof:"README.md; assets/demo.png"
  },
  {
    slug:"sparkwa",name:"SparkWA",eyebrow:"04 / CONNECTED WORKFLOWS",category:"Agents",
    tagline:"A WhatsApp conversation, connected to your Spark agent.",
    description:"A Go-based WhatsApp-to-Gemini Spark bridge that relays instructions through Gmail, with sender verification, local state, and an interactive terminal dashboard.",
    features:["WhatsApp-to-Gmail relay","Authorized sender verification","Native Go executable","Terminal setup and control dashboard"],
    audience:"Personal agent operators and developers",stack:"Go · Gmail API · SQLite · WhatsApp",availability:"Source available",pricing:"MIT-licensed repository; infrastructure costs may apply",
    repo:"https://github.com/krishujeniya/sparkwa",docs:"../../sparkwa-docs.html",
    media:"../../assets/sparkwa-banner.png",mediaAlt:"SparkWA project banner",color:"#9de9c8",proof:"README.md; docs/assets/sparkwa-banner.png"
  },
  {
    slug:"sureword",name:"Sureword",eyebrow:"05 / SPEECH RELIABILITY",category:"Applied AI",
    tagline:"Ask again only when it matters.",
    description:"A voice-to-form workflow for frontline visit records that ranks uncertain fields, asks targeted spoken clarification questions, and requires human review before committing data.",
    features:["Hindi-English dictation workflow","Field-level risk scoring","Selective spoken clarification","Human-reviewed records"],
    audience:"Frontline form and voice-workflow researchers",stack:"Python · Gnani Prisma · Gnani Timbre",availability:"Research project",pricing:"No pricing published",
    repo:"https://github.com/krishujeniya/Sureword",media:"https://raw.githubusercontent.com/krishujeniya/Sureword/main/artifacts/audit/live_desktop_redesign.png",mediaAlt:"Sureword interface audit screenshot",color:"#ffcfad",proof:"README.md; artifacts/audit/live_desktop_redesign.png"
  },
  {
    slug:"seva-setu",name:"Seva Setu",eyebrow:"06 / AGENTIC PUBLIC SERVICES",category:"Agents",
    tagline:"Access to welfare, designed around people.",
    description:"A demonstration welfare portal with separate citizen and officer MCP interfaces, role-bound authentication, and an application system of record for a fictional administrative district.",
    features:["Citizen and officer MCP endpoints","OAuth-protected agent access","Bilingual citizen portal","Administrative workflows"],
    audience:"Public-service prototypes and MCP developers",stack:"Flask · PostgreSQL · MCP · OAuth",availability:"Source available",pricing:"No commercial offer published",
    repo:"https://github.com/krishujeniya/sewasetu-ai",color:"#a9d9ff",proof:"README.md"
  },
  {
    slug:"waferguard",name:"WaferGuard",eyebrow:"07 / INDUSTRIAL INTELLIGENCE",category:"Applied AI",
    tagline:"Find the signal before the defect.",
    description:"An MLOps-oriented semiconductor fault detection project with batch telemetry validation, drift-oriented preprocessing, clustered classification, and prediction reports.",
    features:["Sensor batch validation","Drift and outlier preprocessing","Clustered ensemble classification","Versioned data workflows"],
    audience:"MLOps engineers and manufacturing researchers",stack:"Python · scikit-learn · DVC",availability:"Research project",pricing:"Source project; pricing not published",
    repo:"https://github.com/krishujeniya/WaferGuard",media:"https://raw.githubusercontent.com/krishujeniya/WaferGuard/main/assets/social-card.svg",mediaAlt:"WaferGuard repository social artwork",color:"#97d8ea",proof:"README.md; assets/social-card.svg"
  },
  {
    slug:"vemy",name:"Vemy",eyebrow:"08 / AI ASSISTANT",category:"Applications",
    tagline:"An AI workspace built for conversation.",
    description:"A Next.js AI chat interface with streaming responses, markdown rendering, multimodal input, and progressive-web-app features described in its repository.",
    features:["Streaming chat interface","Voice and image input","Markdown-rich responses","Progressive web app"],
    audience:"People exploring multimodal assistants",stack:"Next.js · React · TypeScript",availability:"Source available",pricing:"No public subscription price verified",
    repo:"https://github.com/krishujeniya/Vemy",media:"https://raw.githubusercontent.com/krishujeniya/Vemy/main/public/LogoWhite.png",mediaAlt:"Vemy application logo",color:"#a4d8ff",proof:"README.md; public/LogoWhite.png"
  },
  {
    slug:"hajree",name:"Hajree",eyebrow:"09 / EDUCATION OPERATIONS",category:"Applied AI",
    tagline:"Attendance with intelligence at its core.",
    description:"An AI-assisted attendance platform with admin, faculty, and student portals, facial recognition, anti-spoofing, and attendance reporting described in the source.",
    features:["Role-based portals","Facial recognition matching","Attendance reporting and alerts","Live or photo-based capture"],
    audience:"Educational institutions and attendance teams",stack:"FastAPI · React · DeepFace",availability:"Source available",pricing:"No commercial pricing published",
    repo:"https://github.com/krishujeniya/Hajree",media:"https://raw.githubusercontent.com/krishujeniya/Hajree/main/assets/logo.png",mediaAlt:"Hajree project logo",color:"#e9d7a3",proof:"README.md; assets/logo.png"
  },
  {
    slug:"wanderlust",name:"Wanderlust",eyebrow:"10 / AI TRAVEL",category:"Applications",
    tagline:"A starting point for trips worth remembering.",
    description:"A lightweight travel landing experience designed around an IBM watsonx Orchestrate chat integration, responsive layout, and dependency-free frontend.",
    features:["Travel-oriented chat interface","IBM watsonx widget integration","Responsive landing experience","Dependency-free frontend"],
    audience:"Travelers and conversational interface builders",stack:"HTML · CSS · JavaScript · IBM watsonx",availability:"Website referenced",pricing:"No paid plan documented",
    repo:"https://github.com/krishujeniya/Wanderlust",website:"https://krishujeniya.github.io/Wanderlust/",
    media:"https://raw.githubusercontent.com/krishujeniya/Wanderlust/main/assets/hero-bg.png",mediaAlt:"Wanderlust website hero image",color:"#d7d5ac",proof:"README.md; assets/hero-bg.png"
  },
  {
    slug:"mostlyoutlier",name:"MostlyOutlier",eyebrow:"11 / AGENT MARKETPLACE",category:"Agents",
    tagline:"Automation built for everyday work.",
    description:"An AI agent marketplace concept featuring prebuilt automation workflows. Its README advertises one-dollar agents; current purchasing availability has not been independently confirmed.",
    features:["AI workflow catalog","n8n-oriented agents","Automation-focused browsing","Marketplace concept"],
    audience:"People exploring practical automations",stack:"Web · AI agents · n8n",availability:"Website referenced",pricing:"Repository advertises $1 agents; current offers unverified",
    repo:"https://github.com/krishujeniya/MostlyOutlier",website:"https://mostlyoutlier.vercel.app/",color:"#e1c6ff",proof:"README.md"
  },
  {
    slug:"snaptext",name:"SnapText OCR",eyebrow:"12 / DOCUMENT INTELLIGENCE",category:"Applied AI",
    tagline:"From an image to usable text.",
    description:"A mobile OCR application for image-to-text extraction with editing, copying, multilingual text support, and a history of past scans.",
    features:["Image to text","Edit and copy results","Scan history","Multilingual support"],
    audience:"Mobile users extracting text from images",stack:"Flutter · Dart · OCR",availability:"Source available",pricing:"No purchase price documented",
    repo:"https://github.com/krishujeniya/snaptext-ocr",media:"https://raw.githubusercontent.com/krishujeniya/snaptext-ocr/main/assets/Screenshot%20from%202024-06-17%2011-33-12.png",mediaAlt:"SnapText OCR screenshot from repository",color:"#c6d5ff",proof:"README.md; assets/Screenshot from 2024-06-17 11-33-12.png"
  },
  {
    slug:"netsweep",name:"NetSweep",eyebrow:"13 / SECURITY TOOLING",category:"Developer tools",
    tagline:"Network reconnaissance, made reproducible.",
    description:"A Bash and Nmap pipeline for multi-stage network reconnaissance, input validation, and timestamped scan-result logs.",
    features:["Validated scan inputs","Multi-stage reconnaissance","Timestamped audit outputs","Shell-first execution"],
    audience:"Network administrators and security labs",stack:"Bash · Nmap",availability:"Source available",pricing:"Repository available; no pricing published",
    repo:"https://github.com/krishujeniya/NetSweep",color:"#a6d4cc",proof:"README.md"
  },
  {
    slug:"fitness-tracker",name:"FitnessTracker",eyebrow:"14 / HUMAN ACTIVITY ML",category:"Applied AI",
    tagline:"Discover patterns in motion.",
    description:"A human activity classification project with sensor-data management, feature engineering, model training, and an MLOps pipeline.",
    features:["Activity classification models","Feature engineering pipeline","Data organization","Containerized workflows"],
    audience:"Wearables and activity recognition researchers",stack:"Python · Machine learning · Docker",availability:"Research project",pricing:"Not commercially offered",
    repo:"https://github.com/krishujeniya/fitness-activity-recognition",color:"#f1cdbb",proof:"README.md"
  },
  {
    slug:"stress-prediction",name:"Stress Level Prediction",eyebrow:"15 / PHYSIOLOGICAL ML",category:"Applied AI",
    tagline:"Interpreting the signals of sleep.",
    description:"A five-class stress classification project using sleep-related biosignals, multiple machine learning models, SHAP diagnostics, and a Streamlit interface.",
    features:["Five-class stress classification","Biosignal feature set","Model comparison","SHAP-based explainability"],
    audience:"ML and biosignal research practitioners",stack:"Python · PyTorch · Streamlit · SHAP",availability:"Research project",pricing:"Not commercially offered",
    repo:"https://github.com/krishujeniya/stress-level-prediction",color:"#bcd6df",proof:"README.md"
  },
  {
    slug:"drug-effectiveness",name:"Drug Effectiveness Predictor",eyebrow:"16 / CLINICAL ML",category:"Applied AI",
    tagline:"Exploring outcomes through machine learning.",
    description:"An experimental treatment response prediction pipeline combining tabular clinical features, gradient boosting, and a Streamlit research dashboard. Not a medical decision tool.",
    features:["Tabular feature processing","Gradient boosting pipeline","Streamlit interface","Clinical-data modeling experiment"],
    audience:"Clinical data science researchers",stack:"Python · scikit-learn · Streamlit",availability:"Research project",pricing:"Not commercially offered",
    repo:"https://github.com/krishujeniya/drug-effectiveness-prediction",color:"#b8d8bd",proof:"README.md"
  },
  {
    slug:"studentportify",name:"StudentPortify",eyebrow:"17 / CAMPUS EXPERIENCE",category:"Applications",
    tagline:"Campus life, in one mobile experience.",
    description:"A student-focused mobile application with profiles, assignments, grades, schedules, announcements, and discussion tools described in its README.",
    features:["Student profiles","Assignment tracking","Course grades and feedback","Campus schedules and discussions"],
    audience:"Students and education application developers",stack:"Flutter · Dart",availability:"Source available",pricing:"No paid plan documented",
    repo:"https://github.com/krishujeniya/studentportify-mobile",media:"https://raw.githubusercontent.com/krishujeniya/studentportify-mobile/main/logo.png",mediaAlt:"StudentPortify logo",color:"#bab9ff",proof:"README.md; logo.png"
  },
  {
    slug:"comune",name:"Comune",eyebrow:"18 / COMMUNICATION",category:"Applications",
    tagline:"A place for messages and connection.",
    description:"A cross-platform chat application repository covering real-time messages, profiles, file and media sharing, and group conversations.",
    features:["Real-time messaging","Profiles and sign-in","Group conversations","Media and file sharing"],
    audience:"Messaging app users and mobile builders",stack:"Flutter · Dart",availability:"Source available",pricing:"No published plan",
    repo:"https://github.com/krishujeniya/comune-chat",media:"https://raw.githubusercontent.com/krishujeniya/comune-chat/main/assets/images/logo.png",mediaAlt:"Comune project logo",color:"#b5d7fb",proof:"README.md; assets/images/logo.png"
  },
  {
    slug:"echonest",name:"EchoNest",eyebrow:"19 / SOCIAL MOBILE",category:"Applications",
    tagline:"Share updates, build connections.",
    description:"A Flutter-based social application covering user accounts, post feeds, and updates with text and media.",
    features:["User profiles","Posts with text and media","Mobile-first social experience","Flutter application source"],
    audience:"Mobile application builders",stack:"Flutter · Dart",availability:"Source available",pricing:"No published plan",
    repo:"https://github.com/krishujeniya/echonest-mobile",media:"https://raw.githubusercontent.com/krishujeniya/echonest-mobile/main/assets/iw.png",mediaAlt:"EchoNest project visual",color:"#ffcfb0",proof:"README.md; assets/iw.png"
  },
  {
    slug:"coursly",name:"Coursly",eyebrow:"20 / LEARNING PLATFORM",category:"Applications",
    tagline:"Learning belongs within reach.",
    description:"A course-learning platform built with PHP and Bootstrap, including a course catalog, lessons, and learner progress features described in the repository.",
    features:["Course catalog","Interactive course content","Learning progress","Certification-oriented workflows"],
    audience:"Learners and educators",stack:"PHP · Bootstrap",availability:"Source available",pricing:"No current price verified",
    repo:"https://github.com/krishujeniya/coursly-lms",media:"https://raw.githubusercontent.com/krishujeniya/coursly-lms/main/assets/img/mylogob.png",mediaAlt:"Coursly project logo",color:"#f3d9a9",proof:"README.md; assets/img/mylogob.png"
  }
]);
export const categories = Object.freeze(["All","Agents","Developer tools","Applied AI","Applications"]);
export function bySlug(slug){return products.find(product=>product.slug===slug);}
