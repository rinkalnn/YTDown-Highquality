export namespace main {
	
	export class AppInfo {
	    name: string;
	    version: string;
	    author: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	    }
	}
	export class AppUpdateInfo {
	    current: string;
	    latest: string;
	    available: boolean;
	    releaseUrl: string;
	    downloadUrl: string;
	    assetName: string;
	    releaseNotes: string;
	
	    static createFrom(source: any = {}) {
	        return new AppUpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.available = source["available"];
	        this.releaseUrl = source["releaseUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.assetName = source["assetName"];
	        this.releaseNotes = source["releaseNotes"];
	    }
	}
	export class BinaryVersion {
	    name: string;
	    current: string;
	    latest: string;
	    canUpgrade: boolean;
	    updatePath: string;
	
	    static createFrom(source: any = {}) {
	        return new BinaryVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.canUpgrade = source["canUpgrade"];
	        this.updatePath = source["updatePath"];
	    }
	}
	export class CookieConfig {
	    mode: string;
	    selected_browser: string;
	    manual_header?: string;
	
	    static createFrom(source: any = {}) {
	        return new CookieConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.selected_browser = source["selected_browser"];
	        this.manual_header = source["manual_header"];
	    }
	}
	export class GalleryDownloadOptions {
	    savePath: string;
	    threads: number;
	    browser: string;
	    ugoiraToWebm: boolean;
	    formats: string[];
	    archive: boolean;
	    extraArgs: string;
	
	    static createFrom(source: any = {}) {
	        return new GalleryDownloadOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.savePath = source["savePath"];
	        this.threads = source["threads"];
	        this.browser = source["browser"];
	        this.ugoiraToWebm = source["ugoiraToWebm"];
	        this.formats = source["formats"];
	        this.archive = source["archive"];
	        this.extraArgs = source["extraArgs"];
	    }
	}
	export class VideoInfo {
	    title: string;
	    thumbnail: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new VideoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.thumbnail = source["thumbnail"];
	        this.id = source["id"];
	    }
	}

}

