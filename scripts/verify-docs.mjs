#!/usr/bin/env node
import assert from "node:assert/strict";
import {readFileSync,existsSync} from "node:fs";
import {execFileSync} from "node:child_process";
import {fileURLToPath} from "node:url";
import path from "node:path";
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"..");
const docs=path.join(root,"docs");
const routes=["","docs/","docs/vps-install/","docs/spark-setup/","docs/gmail-oauth/","docs/security/","docs/troubleshooting/","reference/complete-guide.html"];
const assets=["assets/docs.css","assets/docs.js","assets/search-index.js","assets/sparkwa-mark.svg","assets/sparkwa-banner.png"];
function localFile(route){return path.join(docs,route.endsWith("/")?route+"index.html":route)}
for(const src of ["assets/docs.js","assets/search-index.js"])execFileSync(process.execPath,["--check",path.join(docs,src)],{stdio:"inherit"});
for(const route of routes){assert(existsSync(localFile(route)),"Missing route "+route)}
for(const asset of assets){assert(existsSync(path.join(docs,asset)),"Missing required asset "+asset)}
const internal=new Set,checked=[];
for(const route of routes){
 const file=localFile(route),html=readFileSync(file,"utf8");
 assert(html.includes('<main id="main">')||route.startsWith("reference/"),"Missing main for "+route);
 assert(/<h1\b/.test(html),"Missing H1 for "+route);
 assert(/<meta name="description"/.test(html),"Missing SEO description "+route);
 assert(/<link rel="canonical"/.test(html),"Missing canonical "+route);
 if(route!=="reference/complete-guide.html"){
   assert(html.includes("assets/docs.js"),"Missing enhancement script "+route);
   assert(html.includes('class="skip"'),"Missing skip link "+route);
 }
 const hrefs=[...html.matchAll(/\bhref="([^"]+)"/g)].map(x=>x[1]);
 for(const href of hrefs){
  if(href.startsWith("#"))continue;
  const url=new URL(href,"https://example.invalid/"+route);
  if(url.hostname!=="example.invalid")continue;
  const target=decodeURIComponent(url.pathname.slice(1));
  if(!target){continue}
  if(target==="LICENSE")continue;
  if(!target.startsWith("docs/")&&!target.startsWith("reference/")&&!target.startsWith("assets/"))continue;
  const targetFile=localFile(target);
  assert(existsSync(targetFile),"Broken local link in "+route+": "+href+" -> "+targetFile);
  checked.push(href);
 }
}
const index=readFileSync(path.join(docs,"assets/search-index.js"),"utf8");
for(const route of routes.filter(x=>x.startsWith("docs/")&&x!=="docs/")){
 assert(index.includes('"path":"'+route+'"'),"Search index missing "+route);
}
const home=readFileSync(localFile(""),"utf8");
for(const id of ["architecture","quickstart","guides","faq"]){assert(home.includes('id="'+id+'"'),"Missing homepage section "+id)}
const design=readFileSync(path.join(docs,"DESIGN_SYSTEM.md"),"utf8");
const actualRows=design.split("\n").filter(line=>/^\| (?:[LASCPOMTR]\d+) \|/.test(line)).length;
assert.equal(actualRows,93,"Missing effects audit rows");
const sitemap=readFileSync(path.join(docs,"sitemap.xml"),"utf8");
for(const r of routes)assert(sitemap.includes("https://krishujeniya.github.io/sparkwa/"+r),"Missing sitemap route "+r);
assert(!home.includes("KRISH/LAB"),"Unrelated portfolio branding detected");
assert(!home.includes("MostlyOutlier"),"Unrelated portfolio brand detected");
console.log("PASS: JS syntax, "+routes.length+" HTML routes, "+assets.length+" static assets, "+checked.length+" local links, sitemap, 93 effects rows, and source-backed search index");
