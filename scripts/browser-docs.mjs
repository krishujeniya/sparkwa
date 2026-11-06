#!/usr/bin/env node
import {chromium} from "playwright";
import {mkdirSync,writeFileSync} from "node:fs";
import assert from "node:assert/strict";
const base="http://127.0.0.1:4173/";
const sizes=[320,360,390,768,1024,1280,1440,1920];
const out="docs/audit/site-v2";mkdirSync(out,{recursive:true});
const browser=await chromium.launch({headless:true,...(process.env.CHROME_PATH?{executablePath:process.env.CHROME_PATH}:{}),args:["--no-sandbox"]});
const errors=[],results=[];
async function check(page,label){
 const {scrollWidth,innerWidth}=await page.evaluate(()=>({scrollWidth:document.documentElement.scrollWidth,innerWidth:window.innerWidth}));
 assert(scrollWidth<=innerWidth+2,label+" horizontal overflow: "+scrollWidth+" > "+innerWidth);
 results.push({label,scrollWidth,innerWidth,ok:true});
}
try{
 for(const width of sizes){
  const context=await browser.newContext({viewport:{width,height:900},reducedMotion:"reduce"});
  const page=await context.newPage();
  page.on("pageerror",e=>errors.push(width+": "+e.message));
  await page.goto(base,{waitUntil:"domcontentloaded"});
  await page.screenshot({path:out+"/home-"+width+".png",fullPage:true});
  await check(page,"home "+width);
  await page.goto(base+"docs/vps-install/",{waitUntil:"domcontentloaded"});
  await page.screenshot({path:out+"/install-"+width+".png",fullPage:true});
  await check(page,"install "+width);
  assert((await page.locator("h1").count())===1,"Expected one H1");
  assert(await page.locator(".article h2").count()>0,"Guide sections missing");
  await context.close();
 }
 const ctx=await browser.newContext({viewport:{width:1280,height:900},permissions:["clipboard-read","clipboard-write"]});
 const page=await ctx.newPage();
 page.on("pageerror",e=>errors.push("interactive: "+e.message));
 await page.goto(base,{waitUntil:"domcontentloaded"});
 assert(await page.locator("#mobile-navigation").isHidden(),"Mobile navigation visible at desktop width");
 assert(await page.locator(".topnav").isVisible(),"Desktop navigation not visible");
 assert((await page.locator(".hero-h1").innerText()).includes("Spark agent from WhatsApp"),"Hero headline did not update");
 assert((await page.locator(".hero-process li").count())===3,"Product message flow is missing");
 assert(await page.locator(".hero-process a[href*=spark-setup]").isVisible(),"Setup guide link missing");
 await page.keyboard.press("Control+k");
 await page.locator("#docs-search-box").waitFor();
 await page.locator("#docs-search-box").fill("OAuth");
 assert(await page.locator(".search-result").count()>0,"Search returned no results");
 await page.keyboard.press("Escape");
 assert(await page.locator(".search-overlay").isHidden(),"Search did not close");
 await page.locator("[data-theme-toggle]").click();
 assert(await page.locator("html").getAttribute("data-theme")==="dark","Theme toggle did not persist to DOM");
 await page.reload({waitUntil:"domcontentloaded"});
 assert(await page.locator("html").getAttribute("data-theme")==="dark","Theme toggle preference not persisted");
 await page.locator("#faq details").first().locator("summary").click();
 assert(await page.locator("#faq details").first().evaluate(e=>e.open),"FAQ did not expand");
 await page.goto(base+"docs/gmail-oauth/",{waitUntil:"domcontentloaded"});
 await page.locator(".code-head button").first().click();
 await page.waitForFunction(()=>document.querySelector("#status-announcement")?.textContent?.includes("Copied"),null,{timeout:3000});
 await page.goto(base+"docs/vps-install/",{waitUntil:"domcontentloaded"});
 assert(await page.locator(".docs-toc a").count()>0,"TOC absent");
 const side=await browser.newContext({viewport:{width:390,height:844},reducedMotion:"reduce"});
 const mobile=await side.newPage();
 await mobile.goto(base+"docs/security/",{waitUntil:"domcontentloaded"});
 await mobile.locator("[data-sidebar-toggle]").click();
 assert(await mobile.locator(".docs-sidebar").evaluate(el=>el.classList.contains("open")),"Mobile sidebar failed");
 await mobile.keyboard.press("Escape");
 assert(!(await mobile.locator(".docs-sidebar").evaluate(el=>el.classList.contains("open"))),"Escape did not close sidebar");
 await mobile.goto(base,{waitUntil:"domcontentloaded"});
 assert(await mobile.locator("#mobile-navigation").isHidden(),"Mobile navigation visible before opening");
 await mobile.locator("[data-menu-button]").click();
 assert(await mobile.locator("#mobile-navigation").evaluate(el=>el.classList.contains("open")),"Mobile header menu failed");
 await mobile.keyboard.press("Escape");
 assert(!(await mobile.locator("#mobile-navigation").evaluate(el=>el.classList.contains("open"))),"Mobile menu Escape failed");
 const motion=await mobile.locator(".hero-h1").evaluate(e=>getComputedStyle(e).animationDuration);
 assert(motion==="0s"||motion==="0.01ms","Reduced motion not honored: "+motion);
 assert(await mobile.locator(".hero-process").isVisible(),"Documentation flow missing on mobile");
 assert((await mobile.locator(".hero-process li").count())===3,"Mobile product flow incomplete");
 await side.close();
 const noJS=await browser.newContext({javaScriptEnabled:false});
 const pageNoJS=await noJS.newPage();
 await pageNoJS.goto(base+"docs/security/");
 assert(await pageNoJS.locator(".article h2").count()>0,"No-JS article unavailable");
 await noJS.close();
 const slow=await browser.newContext({viewport:{width:390,height:844}});
 await slow.route("**/assets/sparkwa-mark.svg",async route=>{await new Promise(r=>setTimeout(r,2500));await route.abort()});
 const slowPage=await slow.newPage();
 await slowPage.goto(base,{waitUntil:"domcontentloaded"});
 assert(await slowPage.locator("h1").isVisible(),"Slow media blocked text content");
 await slow.close();
 assert.equal(errors.length,0,"JavaScript errors: "+errors.join("; "));
 writeFileSync(out+"/results.json",JSON.stringify({results,errors,search:true,theme:true,copy:true,faq:true,toc:true,mobileMenu:true,mobileSidebar:true,reducedMotion:true,noJS:true,slowNetwork:true,blockedAutoplay:"not applicable: no videos"},null,2));
 console.log("PASS: 16 viewport screenshots, zero overflow, docs search, theme, copy, FAQ, TOC, mobile navigation, reduced motion, no-JS reading, slow network and no page errors");
}catch(e){console.error("FAILED:",e);writeFileSync(out+"/failure.json",JSON.stringify({errors,message:String(e)},null,2));process.exitCode=1}
finally{await browser.close()}
