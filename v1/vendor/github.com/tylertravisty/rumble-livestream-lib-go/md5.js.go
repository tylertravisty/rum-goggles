package rumblelivestreamlib

const md5 = `
/* @license
 * Version 2.2 Copyright (C) Paul Johnston 1999 - 2009
 * Other contributors: Greg Holt, Andrew Kepert, Ydnar, Lostinet
 * Distributed under the BSD License
 * See http://pajhome.org.uk/crypt/md5 for more info.
 */

md5 = function() {
	function n(){
        this.hex="0123456789abcdef".split("")
    }
    return n.prototype={
        hash:function(n){
            var h=this;
            return h.binHex(h.binHash(h.strBin(n),n.length<<3))
        },
        hashUTF8:function(n){
            return this.hash(this.encUTF8(n))
        },
        hashRaw:function(n){
            var h=this;
            return h.binStr(h.binHash(h.strBin(n),n.length<<3))
        },
        hashRawUTF8:function(n){
            return this.hashRaw(this.encUTF8(n))
        },
        hashStretch:function(n,h,i){
            return this.binHex(this.binHashStretch(n,h,i))
        },
        binHashStretch:function(n,h,i){
            var t,r,f=this,n=f.encUTF8(n),e=h+n,g=32+n.length<<3,o=f.strBin(n),a=o.length,e=f.binHash(f.strBin(e),e.length<<3);
            for(i=i||1024,t=0;t<i;t++){
                for(e=f.binHexBin(e),r=0;r<a;r++)e[8+r]=o[r];e=f.binHash(e,g)
            }
            return e
        },
        encUTF8:function(n){
            for(var h,i,t="",r=0,f=n.length-1;r<=f;)h=n.charCodeAt(r++),r<f&&55296<=h&&h<=56319&&56320<=(i=n.charCodeAt(r))&&i<=57343&&(h=65536+((1023&h)<<10)+(1023&i),r++),h<=127?t+=String.fromCharCode(h):h<=2047?t+=String.fromCharCode(192|h>>>6&31,128|63&h):h<=65535?t+=String.fromCharCode(224|h>>>12&15,128|h>>>6&63,128|63&h):h<=2097151&&(t+=String.fromCharCode(240|h>>>18&7,128|h>>>12&63,128|h>>>6&63,128|63&h));
            return t
        },
        strBin:function(n){
            for(var h=n.length<<3,i=[],t=0;t<h;t+=8)i[t>>5]|=(255&n.charCodeAt(t>>3))<<(31&t);
            return i
        },
        binHex:function(n){
            for(var h,i,t="",r=n.length<<5,f=0;f<r;f+=8)i=(h=n[f>>5]>>>(31&f)&255)>>>4&15,t+=this.hex[i]+this.hex[h&=15];
            return t
        },
        binStr:function(n){
            for(var h,i="",t=n.length<<5,r=0;r<t;r+=8)h=n[r>>5]>>>(31&r)&255,i+=String.fromCharCode(h);
            return i
        },
        binHexBin:function(n){
            for(var h,i,t=n.length<<5,r=[],f=0;f<t;f+=8)i=(h=n[f>>5]>>>(31&f)&255)>>>4&15,r[f>>4]|=(9<i?87:48)+i+((9<(h&=15)?87:48)+h<<8)<<((15&f)<<1);
            return r
        },
        ff:function(n,h,i,t,r,f,e){
            i=h&i|~h&t,t=(65535&n)+(65535&i)+(65535&r)+(65535&e);
            return((i=(i=(n>>16)+(i>>16)+(r>>16)+(e>>16)+(t>>16)<<16|65535&t)<<f|i>>>32-f)>>16)+(h>>16)+((t=(65535&i)+(65535&h))>>16)<<16|65535&t
        },
        gg:function(n,h,i,t,r,f,e){
            i=h&t|i&~t,t=(65535&n)+(65535&i)+(65535&r)+(65535&e);
            return((i=(i=(n>>16)+(i>>16)+(r>>16)+(e>>16)+(t>>16)<<16|65535&t)<<f|i>>>32-f)>>16)+(h>>16)+((t=(65535&i)+(65535&h))>>16)<<16|65535&t
        },
        hh:function(n,h,i,t,r,f,e){
            i=h^i^t,t=(65535&n)+(65535&i)+(65535&r)+(65535&e);
            return((i=(i=(n>>16)+(i>>16)+(r>>16)+(e>>16)+(t>>16)<<16|65535&t)<<f|i>>>32-f)>>16)+(h>>16)+((t=(65535&i)+(65535&h))>>16)<<16|65535&t
        },
        ii:function(n,h,i,t,r,f,e){
            i^=h|~t,t=(65535&n)+(65535&i)+(65535&r)+(65535&e);
            return((i=(i=(n>>16)+(i>>16)+(r>>16)+(e>>16)+(t>>16)<<16|65535&t)<<f|i>>>32-f)>>16)+(h>>16)+((t=(65535&i)+(65535&h))>>16)<<16|65535&t
        },
        binHash:function(n,h){
            var i,t,r,f,e,g,o=1732584193,a=-271733879,u=-1732584194,s=271733878,c=this;for(n[h>>5]|=128<<(31&h),n[14+(h+64>>>9<<4)]=h,i=n.length,t=0;t<i;t+=16)g=o,r=a,f=u,e=s,o=c.ff(o,a,u,s,n[t+0],7,-680876936),s=c.ff(s,o,a,u,n[t+1],12,-389564586),u=c.ff(u,s,o,a,n[t+2],17,606105819),a=c.ff(a,u,s,o,n[t+3],22,-1044525330),o=c.ff(o,a,u,s,n[t+4],7,-176418897),s=c.ff(s,o,a,u,n[t+5],12,1200080426),u=c.ff(u,s,o,a,n[t+6],17,-1473231341),a=c.ff(a,u,s,o,n[t+7],22,-45705983),o=c.ff(o,a,u,s,n[t+8],7,1770035416),s=c.ff(s,o,a,u,n[t+9],12,-1958414417),u=c.ff(u,s,o,a,n[t+10],17,-42063),a=c.ff(a,u,s,o,n[t+11],22,-1990404162),o=c.ff(o,a,u,s,n[t+12],7,1804603682),s=c.ff(s,o,a,u,n[t+13],12,-40341101),u=c.ff(u,s,o,a,n[t+14],17,-1502002290),a=c.ff(a,u,s,o,n[t+15],22,1236535329),o=c.gg(o,a,u,s,n[t+1],5,-165796510),s=c.gg(s,o,a,u,n[t+6],9,-1069501632),u=c.gg(u,s,o,a,n[t+11],14,643717713),a=c.gg(a,u,s,o,n[t+0],20,-373897302),o=c.gg(o,a,u,s,n[t+5],5,-701558691),s=c.gg(s,o,a,u,n[t+10],9,38016083),u=c.gg(u,s,o,a,n[t+15],14,-660478335),a=c.gg(a,u,s,o,n[t+4],20,-405537848),o=c.gg(o,a,u,s,n[t+9],5,568446438),s=c.gg(s,o,a,u,n[t+14],9,-1019803690),u=c.gg(u,s,o,a,n[t+3],14,-187363961),a=c.gg(a,u,s,o,n[t+8],20,1163531501),o=c.gg(o,a,u,s,n[t+13],5,-1444681467),s=c.gg(s,o,a,u,n[t+2],9,-51403784),u=c.gg(u,s,o,a,n[t+7],14,1735328473),a=c.gg(a,u,s,o,n[t+12],20,-1926607734),o=c.hh(o,a,u,s,n[t+5],4,-378558),s=c.hh(s,o,a,u,n[t+8],11,-2022574463),u=c.hh(u,s,o,a,n[t+11],16,1839030562),a=c.hh(a,u,s,o,n[t+14],23,-35309556),o=c.hh(o,a,u,s,n[t+1],4,-1530992060),s=c.hh(s,o,a,u,n[t+4],11,1272893353),u=c.hh(u,s,o,a,n[t+7],16,-155497632),a=c.hh(a,u,s,o,n[t+10],23,-1094730640),o=c.hh(o,a,u,s,n[t+13],4,681279174),s=c.hh(s,o,a,u,n[t+0],11,-358537222),u=c.hh(u,s,o,a,n[t+3],16,-722521979),a=c.hh(a,u,s,o,n[t+6],23,76029189),o=c.hh(o,a,u,s,n[t+9],4,-640364487),s=c.hh(s,o,a,u,n[t+12],11,-421815835),u=c.hh(u,s,o,a,n[t+15],16,530742520),a=c.hh(a,u,s,o,n[t+2],23,-995338651),o=c.ii(o,a,u,s,n[t+0],6,-198630844),s=c.ii(s,o,a,u,n[t+7],10,1126891415),u=c.ii(u,s,o,a,n[t+14],15,-1416354905),a=c.ii(a,u,s,o,n[t+5],21,-57434055),o=c.ii(o,a,u,s,n[t+12],6,1700485571),s=c.ii(s,o,a,u,n[t+3],10,-1894986606),u=c.ii(u,s,o,a,n[t+10],15,-1051523),a=c.ii(a,u,s,o,n[t+1],21,-2054922799),o=c.ii(o,a,u,s,n[t+8],6,1873313359),s=c.ii(s,o,a,u,n[t+15],10,-30611744),u=c.ii(u,s,o,a,n[t+6],15,-1560198380),a=c.ii(a,u,s,o,n[t+13],21,1309151649),o=c.ii(o,a,u,s,n[t+4],6,-145523070),s=c.ii(s,o,a,u,n[t+11],10,-1120210379),u=c.ii(u,s,o,a,n[t+2],15,718787259),a=c.ii(a,u,s,o,n[t+9],21,-343485551),o=(o>>16)+(g>>16)+((g=(65535&o)+(65535&g))>>16)<<16|65535&g,a=(a>>16)+(r>>16)+((g=(65535&a)+(65535&r))>>16)<<16|65535&g,u=(u>>16)+(f>>16)+((g=(65535&u)+(65535&f))>>16)<<16|65535&g,s=(s>>16)+(e>>16)+((g=(65535&s)+(65535&e))>>16)<<16|65535&g;
            return[o,a,u,s]
        }
    },
    new n
}();

hashes = [md5.hash(md5.hashStretch(password, salt0, 128) + salt1),md5.hashStretch(password, salt2, 128), salt1]
`
