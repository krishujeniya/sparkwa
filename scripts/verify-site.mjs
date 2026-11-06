#!/usr/bin/env node
import {readFileSync, existsSync} from "node:fs";
import {execFileSync} from "node:child_process";
import assert from "node:assert/strict";
import path from "node:path";
import {fileURLToPath} from "node:url";
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"..");
const docs=path.join(root,"docs");
const pathFor=x=>path.join(docs,x);
for(const file of ["index.html","sparkwa-docs.html","404.html","assets/ecosystem.js","assets/ecosystem.css","assets/products.js","assets/krish-lab-mark.svg","assets/krish-lab-lockup.svg","assets/krish-lab-cover.svg","DESIGN_SYSTEM.md","PRODUCT_INVENTORY.md","QA_REPORT.md"])assert(existsSync(pathFor(file)),file+" missing");
for(const file of ["assets/ecosystem.js","assets/products.js"])execFileSync(process.execPath,["--check",pathFor(file)],{stdio:"inherit"});
const src=readFileSync(pathFor("assets/products.js"),"utf8");
const matches=[...src.matchAll(/^\s*slug:"([^"]+)"/gm)].map(x=>x[1]);
assert.equal(matches.length,20,"Expected all 20 curated products");
assert.equal(new Set(matches).size,matches.length,"Duplicate product slug");
for(const slug of matches){
 const file=pathFor("products/"+slug+"/index.html");
 assert(existsSync(file),"Missing route for "+slug);
 const html=readFileSync(file,"utf8");
 assert(html.includes('data-slug="'+slug+'"'),"Incorrect product route for "+slug);
 assert(html.includes('data-base="../../"'),"Incorrect relative base for "+slug);
}
const homepage=readFileSync(pathFor("index.html"),"utf8");
for(const name of ['id="featured-track"','id="product-grid"','id="stage-panel"','id="category-filters"','id="faq"','id="site-footer"'])assert(homepage.includes(name),"Missing homepage control: "+name);
const sitemap=readFileSync(pathFor("sitemap.xml"),"utf8");
for(const slug of matches)assert(sitemap.includes("/products/"+slug+"/"),"Missing sitemap URL: "+slug);
const old=readFileSync(pathFor("sparkwa-docs.html"),"utf8");assert(old.includes("Headless WhatsApp"),"Original docs may be missing");
console.log("Site static verification passed: "+matches.length+" product routes, sitemap, key controls, assets, and JS syntax.");
