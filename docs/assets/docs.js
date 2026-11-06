/* SparkWA Docs: progressive enhancement only; source content is static HTML. */
(()=>{'use strict';
 const d=document,root=d.documentElement;
 root.classList.add('js');
 const $=(q,scope=d)=>scope.querySelector(q), $$=(q,scope=d)=>Array.from(scope.querySelectorAll(q));
 const base=root.dataset.base||'./';
 const themeKey='sparkwa_docs_theme';
 const reduced=window.matchMedia('(prefers-reduced-motion: reduce)');
 const $status=d.createElement('div');
 $status.id='status-announcement';$status.setAttribute('role','status');$status.setAttribute('aria-live','polite');
 Object.assign($status.style,{position:'absolute',width:'1px',height:'1px',padding:'0',margin:'-1px',overflow:'hidden',clip:'rect(0,0,0,0)',whiteSpace:'nowrap',border:'0'});
 d.body.append($status);
 let toastTimer;
 function announce(message){$status.textContent='';requestAnimationFrame(()=>{$status.textContent=message})}
 function setTheme(t){root.dataset.theme=t;try{localStorage.setItem(themeKey,t)}catch{};$$('[data-theme-toggle]').forEach(b=>{b.setAttribute('aria-label',t==='dark'?'Switch to light theme':'Switch to dark theme');b.setAttribute('aria-pressed',String(t==='dark'));b.title=t==='dark'?'Light mode':'Dark mode';b.textContent=t==='dark'?'☀':'◐'})}
 try{setTheme(localStorage.getItem(themeKey)==='dark'?'dark':'light')}catch{setTheme('light')}
 $$('[data-theme-toggle]').forEach(b=>b.addEventListener('click',()=>setTheme(root.dataset.theme==='dark'?'light':'dark')));
 const mobile=$('#mobile-navigation'),menu=$('[data-menu-button]');
 function setMobile(open){if(!menu||!mobile)return;mobile.classList.toggle('open',open);menu.setAttribute('aria-expanded',String(open));menu.setAttribute('aria-label',open?'Close menu':'Open menu');if(open)mobile.querySelector('a')?.focus()}
 menu?.addEventListener('click',()=>setMobile(!mobile.classList.contains('open')));
 mobile?.querySelectorAll('a').forEach(a=>a.addEventListener('click',()=>setMobile(false)));
 const sidebar=$('.docs-sidebar'),sideToggle=$('[data-sidebar-toggle]');
 function setSidebar(open){if(!sidebar||!sideToggle)return;sidebar.classList.toggle('open',open);sideToggle.setAttribute('aria-expanded',String(open));if(open)sidebar.querySelector('a')?.focus()}
 sideToggle?.addEventListener('click',()=>setSidebar(!sidebar.classList.contains('open')));
 sidebar?.querySelectorAll('a').forEach(a=>a.addEventListener('click',()=>setSidebar(false)));
 /* Real clipboard success / error, never report success before write resolves. */
 async function copyValue(text,b,doneLabel='Copied'){
   const original=b?.textContent||'Copy';
   if(!navigator.clipboard?.writeText){announce('Clipboard unavailable. Select the text manually.');return}
   try{await navigator.clipboard.writeText(text);announce(doneLabel+' to clipboard');if(b){b.textContent=doneLabel;window.clearTimeout(b._restoreTimer);b._restoreTimer=window.setTimeout(()=>b.textContent=original,1700)}}catch{announce('Could not copy. Please select the text manually.')}
 }
 $$('[data-copy]').forEach(b=>b.addEventListener('click',()=>{const val=b.dataset.copy||'',source=b.dataset.copyTarget?$(b.dataset.copyTarget):null;copyValue(source?.textContent||val,b)}));
 $$('.code-block').forEach(box=>{
   const pre=$('pre',box);if(!pre)return;
   const head=$('.code-head',box);if(!head)return;
   if(!$('button',head)){
     const btn=d.createElement('button');btn.type='button';btn.textContent='Copy code';btn.setAttribute('aria-label','Copy this code example');
     head.append(btn);btn.addEventListener('click',()=>copyValue(pre.textContent,btn));
   }
 });
 $$('[data-copy-logo]').forEach(b=>b.addEventListener('click',async()=>{
   try{const response=await fetch(base+'assets/sparkwa-mark.svg');if(!response.ok)throw Error('logo unavailable');await copyValue(await response.text(),b,'SVG copied')}catch{announce('Logo file unavailable')}
 }));
 /* Docs TOC: update active heading and preserve native anchors */
 const toc=$('.docs-toc');
 if(toc&&'IntersectionObserver' in window){
  const anchors=$$('a[href^="#"]',toc),headings=anchors.map(a=>d.getElementById(decodeURIComponent(a.hash.slice(1)))).filter(Boolean);
  const visible=new Map();
  const ob=new IntersectionObserver(entries=>{
    for(const entry of entries)visible.set(entry.target.id,entry.isIntersecting);
    let selected=null;
    for(const h of headings)if(h.getBoundingClientRect().top<=170)selected=h;
    if(!selected)selected=headings[0]||null;
    anchors.forEach(a=>{if(selected&&a.hash==='#'+selected.id)a.setAttribute('aria-current','true');else a.removeAttribute('aria-current')});
  },{rootMargin:'-100px 0px -72% 0px',threshold:0});
  headings.forEach(h=>ob.observe(h));
 }
 /* Compact, accessible static index search (title + body) */
 const catalog=Array.isArray(window.SPARKWA_SEARCH)?window.SPARKWA_SEARCH:[];
 let overlay,input,results,returnFocus,active=0,matched=[],lastQuery='';
 function safeText(v){return String(v||'').replace(/[<>&"]/g,c=>({'<':'&lt;','>':'&gt;','&':'&amp;','"':'&quot;'}[c]))}
 function absoluteDocPath(p){return base+p.replace(/^\/+/,'')}
 function getResults(term){
  const tokens=term.toLowerCase().trim().split(/\s+/).filter(Boolean).slice(0,8);
  if(!tokens.length)return catalog.slice(0,7);
  return catalog.map(item=>{
    const title=item.title.toLowerCase(),description=(item.description||'').toLowerCase(),body=(item.body||'').toLowerCase();
    let score=0;for(const t of tokens){if(title===t)score+=30;else if(title.includes(t))score+=14;if(description.includes(t))score+=7;let pos=body.indexOf(t);if(pos>=0)score+=3}
    return {item,score}
  }).filter(x=>x.score>0).sort((a,b)=>b.score-a.score||a.item.title.localeCompare(b.item.title)).slice(0,10).map(x=>x.item);
 }
 function renderSearch(){
  matched=getResults(input.value);active=0;
  results.innerHTML=matched.length?matched.map((item,i)=>'<a class="search-result'+(i===0?' is-active':'')+'" data-result="'+i+'" href="'+safeText(absoluteDocPath(item.path))+'"><strong>'+safeText(item.title)+'</strong><small>'+safeText(item.description||'SparkWA documentation')+'</small></a>').join(''):'<div class="search-empty">No results. Try “Gmail”, “installation”, or “security”.</div>';
  results.querySelectorAll('[data-result]').forEach(a=>a.addEventListener('mouseenter',()=>setActive(Number(a.dataset.result))));
 }
 function setActive(i){active=Math.max(0,Math.min(i,matched.length-1));results.querySelectorAll('[data-result]').forEach(a=>a.classList.toggle('is-active',Number(a.dataset.result)===active));results.querySelector('[data-result="'+active+'"]')?.scrollIntoView({block:'nearest'})}
 function closeSearch(){if(!overlay||overlay.hidden)return;overlay.hidden=true;d.body.style.removeProperty('overflow');returnFocus?.focus()}
 function openSearch(){
  if(!catalog.length){announce('Search index not loaded');return}
  if(!overlay){
   overlay=d.createElement('div');overlay.className='search-overlay';overlay.hidden=true;
   overlay.innerHTML='<div class="search-dialog" role="dialog" aria-modal="true" aria-labelledby="search-dialog-title"><div class="search-header"><span aria-hidden="true">⌕</span><label id="search-dialog-title" for="docs-search-box" style="position:absolute;clip-path:inset(100%);width:1px;height:1px">Search SparkWA documentation</label><input id="docs-search-box" type="search" autocomplete="off" placeholder="Search SparkWA documentation…"><button class="search-close" type="button" data-search-close aria-label="Close search dialog">ESC</button></div><div class="search-results" id="docs-search-results" role="listbox" aria-label="Search results"></div><div class="search-footer"><span>↑ ↓ Navigate · Enter Open</span><span>Esc Close</span></div></div>';
   d.body.append(overlay);input=$('#docs-search-box',overlay);results=$('#docs-search-results',overlay);
   overlay.addEventListener('mousedown',e=>{if(e.target===overlay)closeSearch()});
   $('[data-search-close]',overlay).addEventListener('click',closeSearch);
   input.addEventListener('input',renderSearch);
   input.addEventListener('keydown',e=>{
     if(e.key==='ArrowDown'){e.preventDefault();setActive(active+1)}
     else if(e.key==='ArrowUp'){e.preventDefault();setActive(active-1)}
     else if(e.key==='Enter'&&matched[active]){e.preventDefault();location.href=absoluteDocPath(matched[active].path)}
   });
   overlay.addEventListener('keydown',e=>{if(e.key!=='Tab')return;const focusables=$$('a[href],button:not([disabled]),input:not([disabled])',overlay);const first=focusables[0],last=focusables[focusables.length-1];if(e.shiftKey&&d.activeElement===first){last.focus();e.preventDefault()}else if(!e.shiftKey&&d.activeElement===last){first.focus();e.preventDefault()}});
  }
  returnFocus=d.activeElement;overlay.hidden=false;d.body.style.overflow='hidden';input.value='';renderSearch();input.focus();
 }
 $$('[data-search-trigger]').forEach(b=>b.addEventListener('click',openSearch));
 d.addEventListener('keydown',e=>{
  const target=e.target,editing=target?.matches?.('input,textarea,[contenteditable=true]');
  if((e.metaKey||e.ctrlKey)&&e.key.toLowerCase()==='k'){e.preventDefault();openSearch()}
  else if(e.key==='Escape'){if(overlay&&!overlay.hidden)closeSearch();else if(sidebar?.classList.contains('open')){setSidebar(false);sideToggle?.focus()}else if(mobile?.classList.contains('open')){setMobile(false);menu?.focus()}}
  else if(e.key==='/'&&!editing&&(!overlay||overlay.hidden)){e.preventDefault();openSearch()}
 });

 /* Scroll state is decorative and never gates reading or click targets. */
 const bar=$('.topbar');
 const progress=d.createElement('div');
 progress.className='reading-progress';
 progress.setAttribute('aria-hidden','true');
 bar?.after(progress);
 let scrollFrame=0;
 function syncScroll(){
   scrollFrame=0;
   const y=window.scrollY;
   bar?.classList.toggle('is-scrolled',y>10);
   if(!reduced.matches){
     const total=Math.max(1,root.scrollHeight-window.innerHeight);
     const ratio=Math.max(0,Math.min(1,y/total));
     progress.style.transform='scaleX('+ratio.toFixed(4)+')';
   }
 }
 function scheduleScroll(){if(!scrollFrame)scrollFrame=requestAnimationFrame(syncScroll)}
 window.addEventListener('scroll',scheduleScroll,{passive:true});
 window.addEventListener('resize',scheduleScroll,{passive:true});
 reduced.addEventListener?.('change',scheduleScroll);
 syncScroll();

 /* Meaningful, one-shot editorial reveals on the marketing homepage only.
    Docs pages are intentionally never hidden, even while JavaScript is loading. */
 if($('.home-hero')&&'IntersectionObserver' in window&&!reduced.matches){
   const reveal=$$('.section-heading,.flow-step,.code-feature,.doc-card,.section .accordion');
   for(const el of reveal){
     el.classList.add('reveal-on-scroll');
     if(el.matches('.flow-step,.doc-card')){
       const index=[...el.parentElement.children].indexOf(el);
       el.style.setProperty('--reveal-delay',Math.min(4,index)*70+'ms');
     }
   }
   const observer=new IntersectionObserver(entries=>{
     for(const entry of entries){
       if(entry.isIntersecting){
         entry.target.classList.add('is-in-view');
         observer.unobserve(entry.target);
       }
     }
   },{rootMargin:'0px 0px -35px 0px',threshold:0.08});
   root.classList.add('motion-ready');
   reveal.forEach(el=>observer.observe(el));
 }

 /* Keep the mobile drawer closed when switching to desktop, and dismiss it
    on outside interaction without disrupting ordinary document scrolling. */
 const desktopQuery=window.matchMedia('(min-width:901px)');
 desktopQuery.addEventListener?.('change',event=>{
   if(event.matches){setMobile(false);setSidebar(false)}
 });
 d.addEventListener('pointerdown',event=>{
   if(mobile?.classList.contains('open')&&!mobile.contains(event.target)&&!menu?.contains(event.target))setMobile(false);
   if(sidebar?.classList.contains('open')&&!sidebar.contains(event.target)&&!sideToggle?.contains(event.target))setSidebar(false);
 });
 /* Failed assets are never replaced with misleading images. */
 $$('img').forEach(img=>img.addEventListener('error',()=>{img.style.display='none';announce('An optional image could not be loaded. The documentation is still available.')},{once:true}));
})();
