export namespace config {
	
	export class Config_Controller_SDLMap_Control {
	    kind: string;
	    index: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Config_Controller_SDLMap_Control(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.index = source["index"];
	        this.name = source["name"];
	    }
	}
	export class Config_Controller_SDLMap {
	    name: string;
	    usb_id: string;
	    data: Config_Controller_SDLMap_Control[];
	
	    static createFrom(source: any = {}) {
	        return new Config_Controller_SDLMap(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.usb_id = source["usb_id"];
	        this.data = this.convertValues(source["data"], Config_Controller_SDLMap_Control);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class Interop_Cab_ControlState_Control {
	    Identifier: string;
	    PropertyName: string;
	    CurrentValue: number;
	    CurrentNormalizedValue: number;
	
	    static createFrom(source: any = {}) {
	        return new Interop_Cab_ControlState_Control(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Identifier = source["Identifier"];
	        this.PropertyName = source["PropertyName"];
	        this.CurrentValue = source["CurrentValue"];
	        this.CurrentNormalizedValue = source["CurrentNormalizedValue"];
	    }
	}
	export class Interop_Cab_ControlState {
	    Name: string;
	    Controls: Interop_Cab_ControlState_Control[];
	
	    static createFrom(source: any = {}) {
	        return new Interop_Cab_ControlState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Controls = this.convertValues(source["Controls"], Interop_Cab_ControlState_Control);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Interop_ChangeEvent {
	    UniqueID: string;
	    DeviceID: string;
	    ControlName: string;
	    Value: number;
	
	    static createFrom(source: any = {}) {
	        return new Interop_ChangeEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UniqueID = source["UniqueID"];
	        this.DeviceID = source["DeviceID"];
	        this.ControlName = source["ControlName"];
	        this.Value = source["Value"];
	    }
	}
	export class Interop_ControllerCalibration_Control {
	    Kind: string;
	    Index: number;
	    Name: string;
	    Min: number;
	    Max: number;
	    Idle: number;
	    Deadzone: number;
	    EasingCurve: number[];
	    Invert: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Interop_ControllerCalibration_Control(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Kind = source["Kind"];
	        this.Index = source["Index"];
	        this.Name = source["Name"];
	        this.Min = source["Min"];
	        this.Max = source["Max"];
	        this.Idle = source["Idle"];
	        this.Deadzone = source["Deadzone"];
	        this.EasingCurve = source["EasingCurve"];
	        this.Invert = source["Invert"];
	    }
	}
	export class Interop_ControllerCalibration {
	    Name: string;
	    DeviceID: string;
	    Controls: Interop_ControllerCalibration_Control[];
	
	    static createFrom(source: any = {}) {
	        return new Interop_ControllerCalibration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.DeviceID = source["DeviceID"];
	        this.Controls = this.convertValues(source["Controls"], Interop_ControllerCalibration_Control);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Interop_ControllerConfiguration {
	    Calibration: Interop_ControllerCalibration;
	    SDLMapping: config.Config_Controller_SDLMap;
	
	    static createFrom(source: any = {}) {
	        return new Interop_ControllerConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Calibration = this.convertValues(source["Calibration"], Interop_ControllerCalibration);
	        this.SDLMapping = this.convertValues(source["SDLMapping"], config.Config_Controller_SDLMap);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Interop_GenericController {
	    UniqueID: string;
	    DeviceID: string;
	    Name: string;
	    IsConfigured: boolean;
	    IsVirtual: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Interop_GenericController(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UniqueID = source["UniqueID"];
	        this.DeviceID = source["DeviceID"];
	        this.Name = source["Name"];
	        this.IsConfigured = source["IsConfigured"];
	        this.IsVirtual = source["IsVirtual"];
	    }
	}
	export class Interop_Profile_Metadata {
	    Path: string;
	    IsEmbedded: boolean;
	    UpdatedAt: string;
	    Warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new Interop_Profile_Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.IsEmbedded = source["IsEmbedded"];
	        this.UpdatedAt = source["UpdatedAt"];
	        this.Warnings = source["Warnings"];
	    }
	}
	export class Interop_Profile {
	    Id: string;
	    Name: string;
	    DeviceID: string;
	    AutoSelect?: boolean;
	    Apps?: string[];
	    Metadata: Interop_Profile_Metadata;
	
	    static createFrom(source: any = {}) {
	        return new Interop_Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.DeviceID = source["DeviceID"];
	        this.AutoSelect = source["AutoSelect"];
	        this.Apps = source["Apps"];
	        this.Metadata = this.convertValues(source["Metadata"], Interop_Profile_Metadata);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Interop_RawEvent {
	    UniqueID: string;
	    DeviceID: string;
	    Kind: string;
	    Index: number;
	    Value: number;
	    Timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new Interop_RawEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UniqueID = source["UniqueID"];
	        this.DeviceID = source["DeviceID"];
	        this.Kind = source["Kind"];
	        this.Index = source["Index"];
	        this.Value = source["Value"];
	        this.Timestamp = source["Timestamp"];
	    }
	}
	export class Interop_SaveControlMapping {
	    ProfileJSON: string;
	    ExistingPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Interop_SaveControlMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProfileJSON = source["ProfileJSON"];
	        this.ExistingPath = source["ExistingPath"];
	    }
	}
	export class Interop_SelectedProfileInfo {
	    Id: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new Interop_SelectedProfileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	    }
	}
	export class Interop_SharedProfile_Author {
	    Name: string;
	    Url?: string;
	
	    static createFrom(source: any = {}) {
	        return new Interop_SharedProfile_Author(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Url = source["Url"];
	    }
	}
	export class Interop_SharedProfile {
	    Name: string;
	    DeviceID: string;
	    Url: string;
	    AutoSelect?: boolean;
	    Apps?: string[];
	    ContainsCalibration?: boolean;
	    Author?: Interop_SharedProfile_Author;
	
	    static createFrom(source: any = {}) {
	        return new Interop_SharedProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.DeviceID = source["DeviceID"];
	        this.Url = source["Url"];
	        this.AutoSelect = source["AutoSelect"];
	        this.Apps = source["Apps"];
	        this.ContainsCalibration = source["ContainsCalibration"];
	        this.Author = this.convertValues(source["Author"], Interop_SharedProfile_Author);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

