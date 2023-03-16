// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {api} from '../models';

export function AddTorrentByHash(arg1:string):Promise<string>;

export function CheckHeader(arg1:string):Promise<boolean>;

export function GetFiles(arg1:string):Promise<Array<api.File>>;

export function GetTorrents():Promise<Array<api.Torrent>>;

export function OpenFolder(arg1:string):Promise<void>;

export function RemoveTorrent(arg1:string,arg2:boolean):Promise<string>;

export function SetActive(arg1:string,arg2:boolean):Promise<string>;
