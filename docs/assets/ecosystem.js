import {products, categories, bySlug} from "./products.js";
const d=document;
const base=d.documentElement.dataset.base||"./";
const $=(s,root=d)=>root.querySelector(s);
const $$=(s,root=d)=>Array.from(root.querySelectorAll(s));
const esc=v=>String(v??"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
const productPath=p=>base+"products/"+p.slug+"/";
const mediaFor=p=>p.slug==="sparkwa"?base+"assets/sparkwa-banner.png":p.media;
const documentationFor=p=>p.slug==="sparkwa"?base+"sparkwa-docs.html":p.docs;
const logoMode=p=>/logo|social-card|agyos-logo|banner/i.test(p.media||"");
const prefersReduced=d.defaultView.matchMedia("(prefers-reduced-motion: reduce)");
function external(url,label,cls="inline-link"){
  return url?'<a class="'+cls+'" href="'+esc(url)+'" target="_blank" rel="noopener noreferrer">'+esc(label)+' <span aria-hidden="true">↗</span></a>':"";
}
function brand(){
  return '<a class="brand" href="'+base+'" aria-label="Krish Lab — home"><img src="'+base+'assets/krish-lab-mark.svg" alt="" width="33" height="33"><span class="brand-word">KRISH<span style="color:#9ebbf4">/</span>LAB<small>INDEPENDENT BUILDS</small></span></a>';
}
function chrome(){
 const header=$("#site-header");
 if(header)header.innerHTML='<div class="shell nav-shell">'+brand()+'<nav class="nav-links" id="primary-navigation" aria-label="Main navigation"><a href="'+base+'#work">Selected work</a><a href="'+base+'#approach">Approach</a><a href="'+base+'#explore">All projects</a><a href="'+base+'#stories">Field notes</a></nav><a class="nav-cta" href="https://github.com/krishujeniya" target="_blank" rel="noopener noreferrer">Open GitHub ↗</a><button class="menu-button" type="button" aria-label="Open navigation" aria-expanded="false" aria-controls="primary-navigation"><span class="menu-icon" aria-hidden="true"><i></i><i></i></span></button></div>';
 const menu=$(".menu-button");
 if(menu){menu.addEventListener("click",()=>{const nav=$("#primary-navigation");const open=!nav.classList.contains("open");nav.classList.toggle("open",open);menu.setAttribute("aria-expanded",String(open));menu.setAttribute("aria-label",open?"Close navigation":"Open navigation")});$$("#primary-navigation a").forEach(a=>a.addEventListener("click",()=>{menu.setAttribute("aria-expanded","false");menu.setAttribute("aria-label","Open navigation");$("#primary-navigation").classList.remove("open")}));}
 const footer=$("#site-footer");
 if(footer)footer.innerHTML='<div class="shell"><div class="footer-top"><div>'+brand()+'<div class="footer-word">Stay <span>curious.</span></div><p class="body-text" style="max-width:370px">Independent explorations in agents, AI, developer tooling, and products. Built by Krish Ujeniya.</p></div><div class="footer-links"><h4>Explore</h4><a href="'+base+'#work">Selected work</a><a href="'+base+'#explore">Project directory</a><a href="'+base+'#stories">Stories</a><a href="'+base+'#faq">FAQ</a></div><div class="footer-links"><h4>Elsewhere</h4><a href="https://github.com/krishujeniya" target="_blank" rel="noopener noreferrer">GitHub ↗</a><a href="https://github.com/krishujeniya/sparkwa" target="_blank" rel="noopener noreferrer">SparkWA source ↗</a><a href="'+base+'sparkwa-docs.html">SparkWA documentation</a><a href="https://github.com/krishujeniya/sparkwa/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">Repository license ↗</a></div></div><div class="footer-bottom"><span>© '+new Date().getFullYear()+' Krish Ujeniya · Built with care, not a template.</span><span>All media credited to their source repositories · <a href="#top">Back to top ↑</a></span></div></div>';
}
function mediaMarkup(p,extra=""){
 const media=mediaFor(p);
 const text='<span class="media-fallback" aria-hidden="true">'+esc(p.name)+'</span>';
 return text+(media?'<img data-project-image class="'+(logoMode(p)?"logo-media ":"")+esc(extra)+'" src="'+esc(media)+'" alt="'+esc(p.mediaAlt||p.name+" project media")+'" loading="lazy" decoding="async">':"");
}
function watchImages(root=d){
 $$("img[data-project-image]",root).forEach(img=>{if(img.dataset.watched)return;img.dataset.watched="true";img.addEventListener("error",()=>{img.remove()},{once:true})});
}
function featureMarkup(p,i){
 return '<a class="feature-card" href="'+productPath(p)+'" aria-label="Explore '+esc(p.name)+'"><div class="feature-media" style="--tint:'+esc(p.color)+'42">'+mediaMarkup(p)+'</div><div class="feature-foot"><div><span class="eyebrow">'+esc(p.eyebrow)+'</span><h3>'+esc(p.name)+'</h3><p>'+esc(p.tagline)+'</p></div><span class="side-arrow" aria-hidden="true">↗</span></div></a>';
}
function tileMarkup(p,i){
 return '<a class="product-tile" href="'+productPath(p)+'"><div class="tile-visual" style="--tint:'+esc(p.color)+'4f">'+mediaMarkup(p)+'<span class="tile-number">'+String(i+1).padStart(2,"0")+'</span></div><div class="tile-info"><div class="tile-top"><span class="tile-category">'+esc(p.category)+'</span><span class="status">'+esc(p.availability)+'</span></div><h3>'+esc(p.name)+'</h3><p>'+esc(p.tagline)+'</p><span class="tile-link">Explore project <span aria-hidden="true">↗</span></span></div></a>';
}
function renderHome(){
 const featured=$("#featured-track");
 if(featured){
 const arr=products.slice(0,8);
 featured.innerHTML=arr.map(featureMarkup).join("");
 featured.addEventListener("keydown",e=>{if(e.key==="ArrowRight"||e.key==="ArrowLeft"){e.preventDefault();carouselMove(e.key==="ArrowRight"?1:-1)}});
 const prev=$("#carousel-prev"),next=$("#carousel-next");
 function carouselMove(dir){const card=$(".feature-card",featured);if(!card)return;featured.scrollBy({left:(card.getBoundingClientRect().width+24)*dir,behavior:prefersReduced.matches?"instant":"smooth"})}
 prev?.addEventListener("click",()=>carouselMove(-1));next?.addEventListener("click",()=>carouselMove(1));
 const update=()=>{prev.disabled=featured.scrollLeft<=4;next.disabled=featured.scrollLeft+featured.clientWidth>=featured.scrollWidth-6};
 featured.addEventListener("scroll",update,{passive:true});window.addEventListener("resize",update,{passive:true});requestAnimationFrame(update);
 }
 const slides=[products[0],products[2],products[4]];
 const heroImg=$("#hero-image"),heroName=$("#hero-project"),heroKind=$("#hero-kind");
 const pins=$("#hero-pins");
 if(heroImg&&pins){
 pins.innerHTML=slides.map((p,i)=>'<button type="button" aria-label="Show '+esc(p.name)+' project media" aria-pressed="'+(i===0)+'" data-slide="'+i+'"></button>').join("");
 $$("#hero-pins button").forEach(button=>button.addEventListener("click",()=>{const index=Number(button.dataset.slide),p=slides[index];heroImg.classList.add("is-changing");const change=()=>{heroImg.src=mediaFor(p);heroImg.alt=p.mediaAlt||p.name;heroName.textContent=p.name;heroKind.textContent=p.category;$$("#hero-pins button").forEach(b=>b.setAttribute("aria-pressed",String(b===button)));heroImg.classList.remove("is-changing")};if(prefersReduced.matches)change();else setTimeout(change,190)}));
 heroImg.addEventListener("error",()=>{heroImg.style.opacity="0";heroName.textContent="Source media unavailable"});heroImg.addEventListener("load",()=>{heroImg.style.opacity=""});
 }
 const filters=$("#category-filters"),search=$("#product-search"),grid=$("#product-grid"),count=$("#results-meta");
 let category="All",term="";
 const updateGrid=()=>{
   const hits=products.filter(p=>(category==="All"||p.category===category)&&[p.name,p.category,p.tagline,p.description,...p.features].join(" ").toLowerCase().includes(term));
   grid.innerHTML=hits.length?hits.map(tileMarkup).join(""):'<div class="empty-results" style="grid-column:1/-1"><h3>No projects found</h3><p>Try a different search or choose All.</p><button type="button" class="btn btn-outline" id="reset-filters">Show all projects</button></div>';
   count.textContent=hits.length+" of "+products.length+" verified projects";
   $("#reset-filters")?.addEventListener("click",()=>{category="All";term="";search.value="";paintFilters();updateGrid()});
   watchImages(grid);
 };
 const paintFilters=()=>{
   filters.innerHTML=categories.map(c=>'<button type="button" class="filter" aria-pressed="'+(category===c)+'" data-category="'+esc(c)+'">'+esc(c)+'</button>').join("");
   $$("#category-filters button").forEach(b=>b.addEventListener("click",()=>{category=b.dataset.category;paintFilters();updateGrid()}));
 };
 paintFilters();updateGrid();
 search?.addEventListener("input",()=>{term=search.value.trim().toLowerCase();updateGrid()});
 const stages=[
   {name:"Discover",tag:"01 / FIND THE RIGHT QUESTION",headline:"Start with the problem, not the hype.",copy:"The collection spans real engineering, research, productivity, and public-service problems. Browse a verified repository before choosing a direction.",chain:["Explore","Read evidence","Choose a project"]},
   {name:"Build",tag:"02 / DESIGN WITH INTENT",headline:"Make useful systems, then test their edges.",copy:"Projects such as Hadron, Sureword, and WaferGuard document distinct implementation approaches. They are independent projects, not a single connected platform.",chain:["Source","Architecture","Implementation"]},
   {name:"Deploy",tag:"03 / PUT WORK INTO THE WORLD",headline:"Real interfaces. Reproducible code.",copy:"Some projects provide hosted experiences, others provide installation steps or open repositories. Availability is shown individually—never assumed.",chain:["Verify","Run locally","Share your work"]}
 ];
 const tabs=$("#stage-tabs"),panel=$("#stage-panel");
 function chooseStage(index){
  const s=stages[index];
  $$("#stage-tabs button").forEach((b,i)=>b.setAttribute("aria-selected",String(i===index)));
  panel.innerHTML='<span class="smalltag">'+esc(s.tag)+'</span><div><h3>'+esc(s.headline)+'</h3><p>'+esc(s.copy)+'</p></div><div class="stage-chain">'+s.chain.map((a,i)=>(i?'<i aria-hidden="true">→</i>':"")+'<b>'+esc(a)+'</b>').join("")+'</div>';
 }
 tabs?.addEventListener("click",e=>{const b=e.target.closest("button[data-stage]");if(b)chooseStage(Number(b.dataset.stage))});
 tabs?.addEventListener("keydown",e=>{if(!["ArrowUp","ArrowDown","ArrowLeft","ArrowRight","Home","End"].includes(e.key))return;let n=$$("#stage-tabs button").findIndex(x=>x.getAttribute("aria-selected")==="true");if(e.key==="Home")n=0;else if(e.key==="End")n=2;else n=(n+(e.key==="ArrowUp"||e.key==="ArrowLeft"?2:1))%3;chooseStage(n);$$("#stage-tabs button")[n].focus();e.preventDefault()});
 chooseStage(0);
 const stories=$("#story-grid");
 if(stories)stories.innerHTML=[
   [products[0],"An actual dashboard, not a mockup.","Hadron / visual audit"],
   [products[4],"Selective clarification, in the interface.","Sureword / interface audit"],
   [products[2],"An agent runtime you can inspect.","AmberClaw / repo demo"],
   [products[9],"A travel experience built for an agent.","Wanderlust / landing artwork"]
 ].map(([p,title,caption])=>'<a class="story" href="'+productPath(p)+'" style="background:radial-gradient(ellipse at 60% 40%,'+esc(p.color)+'4d,#111722 75%)">'+(mediaFor(p)?'<img data-project-image src="'+esc(mediaFor(p))+'" alt="'+esc(p.mediaAlt||p.name)+'" loading="lazy" decoding="async">':'')+'<div class="story-copy"><small>'+esc(caption)+'</small><h3>'+esc(title)+'</h3><span class="inline-link">See the project ↗</span></div></a>').join("");
 watchImages();
}
function renderProduct(){
 const slug=d.body.dataset.slug||location.pathname.split("/").filter(Boolean).at(-1);
 const p=bySlug(slug),app=$("#product-app");
 if(!app)return;
 if(!p){app.innerHTML='<section class="detail-hero"><div class="shell"><h1 class="detail-title">Not found.</h1><p>That project is not in the verified directory.</p><a class="btn" href="'+base+'#explore">Explore projects ↗</a></div></section>';return;}
 d.documentElement.style.setProperty("--detail-tint",p.color+"45");
 const siteLink=p.website?external(p.website,"Visit referenced site","btn"):"";
 const docsLink=documentationFor(p)?(/^\.\.\/|^https?:/.test(documentationFor(p))?documentationFor(p):documentationFor(p)):"";
 const related=products.filter(x=>x.category===p.category&&x.slug!==p.slug).slice(0,3);
 if(related.length<3) related.push(...products.filter(x=>x.slug!==p.slug&&!related.includes(x)).slice(0,3-related.length));
 app.innerHTML='<section class="detail-hero"><div class="shell"><a class="detail-back" href="'+base+'#explore">← All projects</a><div class="eyebrow">'+esc(p.eyebrow)+'</div><h1 class="detail-title">'+esc(p.name)+'</h1><p class="detail-tagline">'+esc(p.tagline)+'</p><div class="detail-meta"><span class="detail-pill">'+esc(p.category)+'</span><span class="detail-pill">'+esc(p.availability)+'</span></div><div class="detail-actions">'+external(p.repo,"View source","btn")+(p.website?' '+siteLink:"")+(docsLink?' <a class="btn btn-outline" href="'+esc(docsLink)+'"'+(docsLink.startsWith("https")?' target="_blank" rel="noopener noreferrer"':"")+'>Documentation <span aria-hidden="true">↗</span></a>':"")+'</div></div></section>'+
 '<section class="detail-media" aria-label="'+esc(p.name)+' project media" style="--detail-tint:'+esc(p.color)+'3c"><div class="media-fallback" aria-hidden="true">'+esc(p.name)+'</div>'+(mediaFor(p)?'<img data-project-image src="'+esc(mediaFor(p))+'" alt="'+esc(p.mediaAlt||p.name)+'" decoding="async">':"")+'<span class="real-media-note">'+(mediaFor(p)?"Media sourced from the linked project repository":"No verified screenshot available for this project")+'</span></section>'+
 '<section class="shell detail-content"><div><div class="eyebrow">THE PROJECT / OVERVIEW</div><h2>What it sets out <span class="serif">to do.</span></h2><p>'+esc(p.description)+'</p><div class="detail-proof">Source basis: '+esc(p.proof)+'. Project capabilities reflect repository documentation, not independent production certification.</div></div><dl class="detail-facts"><div><dt>Who it is for</dt><dd>'+esc(p.audience)+'</dd></div><div><dt>Technology</dt><dd>'+esc(p.stack)+'</dd></div><div><dt>Availability</dt><dd>'+esc(p.availability)+'</dd></div><div><dt>Access / pricing</dt><dd>'+esc(p.pricing)+'</dd></div><div><dt>Source repository</dt><dd>'+external(p.repo,"Open on GitHub")+'</dd></div></dl></section>'+
 '<section class="detail-features"><div class="shell"><div class="eyebrow">CAPABILITIES / FROM THE SOURCE</div><h2>The pieces that <span class="serif">matter.</span></h2><div class="feature-list">'+p.features.map((f,i)=>'<article class="feature-item"><span>'+String(i+1).padStart(2,"0")+' / '+esc(p.category)+'</span><h3>'+esc(f)+'</h3></article>').join("")+'</div></div></section>'+
 '<section class="detail-related shell"><div class="section-kicker"><span></span>KEEP EXPLORING</div><h2 class="section-title">Other directions<span class="serif">.</span></h2><div class="related-grid">'+related.map(r=>'<a class="related-tile" href="'+productPath(r)+'"><small>'+esc(r.category)+'</small><strong>'+esc(r.name)+'</strong><span>Explore project ↗</span></a>').join("")+'</div></section>';
 watchImages(app);
}
function motion(){
 if(prefersReduced.matches||!("IntersectionObserver" in window)){$$(".reveal").forEach(el=>el.classList.add("is-visible"));return}
 const observer=new IntersectionObserver(entries=>entries.forEach(entry=>{if(entry.isIntersecting){entry.target.classList.add("is-visible");observer.unobserve(entry.target)}}),{threshold:.1,rootMargin:"0px 0px -35px 0px"});
 $$(".reveal").forEach(el=>observer.observe(el));
}
chrome();
if(d.body.dataset.page==="product")renderProduct();else renderHome();
motion();
