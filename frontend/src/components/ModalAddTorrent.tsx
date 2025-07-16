import React, {Component} from 'react';
import {
    AddTorrentByHash,
    AddTorrentByMeta,
    CheckHeader,
    GetFiles,
    RemoveTorrent,
    StartDownload
} from "../../wailsjs/go/main/App";
import {Refresh} from "./Table";
import {Modal} from "./Modal";
import Upload from "../assets/images/icons/upload.svg";
import FileLight from "../../public/light/file-popup.svg";
import FileDark from "../../public/dark/file-popup.svg";

interface State {
    selectFilesStage: boolean
    fieldHash?: string
    fileName?: string
    fieldMeta?: ArrayBuffer
    err: string

    canContinue: boolean
    hash?: string
    files: any[]
}

interface AddTorrentModalProps {
    onExit: () => void
    openHash?: string
    isDark: boolean
}

export class AddTorrentModal extends Component<AddTorrentModalProps, State> {
    constructor(props: AddTorrentModalProps, state: State) {
        super(props, state);

        this.state = {
            selectFilesStage: this.props.openHash ? true : false,
            hash: this.props.openHash,
            err: "",
            files: [],
            canContinue: false,
        }
    }
    inter?: number

    componentDidMount() {
        if (this.props.openHash) {
            this.startCheckFiles(this.props.openHash);
        }
    }

    cancel = () => {
        if (this.state.hash) {
            RemoveTorrent(this.state.hash, true, true).then(Refresh)
        }
        this.props.onExit()
    }

    startCheckFiles = (hash: string) => {
        // check header availability every 50 ms, and load files list when we get it
        this.inter = window.setInterval(() => {
            CheckHeader(hash).then(has => {
                if (has) {
                    clearInterval(this.inter);
                    GetFiles(hash).then((tree) => {
                        this.setState((current) => ({...current, files: tree, canContinue: true}))
                    }).catch((err) => {
                        console.log(err);
                        this.props.onExit();
                    })
                }
            })
        }, 50)
    }

    next = () => {
        if (!this.state.selectFilesStage) {
            let process = (hash: string, err: string) => {
                if (err == "") {
                    this.setState((current) => ({...current, hash, selectFilesStage: true}))
                    this.startCheckFiles(hash);
                    return
                }
                this.setState((current) => ({...current, err, canContinue: true}))
            }

            if (this.state.fieldMeta) {
                // to base64
                const meta = btoa(String.fromCharCode(...new Uint8Array(this.state.fieldMeta)));
                AddTorrentByMeta(meta).then((ti: any) => {
                    process(ti.Hash, ti.Err);
                })
            } else if (this.state.fieldHash) {
                let hash = this.state.fieldHash;
                AddTorrentByHash(hash).then((err) => {
                    process(hash, err);
                })
            }
            this.setState((current) => ({ ...current, canContinue: false }))
        } else {
            let toDownload: string[] = [];
            for (let file of document.getElementsByClassName("file-to-download")) {
                if ((file as HTMLInputElement).checked) {
                    toDownload.push(file.id.slice("file_".length))
                }
            }

            if (toDownload.length > 0) {
                StartDownload(this.state.hash!, toDownload).then()
                this.props.onExit()
            } else {
                this.cancel()
            }
        }
    }

    componentWillUnmount() {
        if (this.inter)
            clearInterval(this.inter)
    }

    checkAndSet = (id: string) => {
        let numChecked = 0;
        let numNotChecked = 0;
        let dep = document.getElementById("dir_"+id)!;
        for (const el of dep.getElementsByClassName("checkbox-file")) {
            (el.children.item(1) as HTMLInputElement)!.checked ? numChecked++ : numNotChecked++;
        }

        let lab = document.getElementById("lab_"+id)!;
        if (numChecked > 0 && numNotChecked == 0) {
            (lab.children.item(1) as HTMLInputElement)!.checked = true;
        } else if (numChecked == 0 && numNotChecked > 0) {
            (lab.children.item(1) as HTMLInputElement)!.checked = false;
        } else {
            (lab.children.item(1) as HTMLInputElement)!.checked = true;
        }

        let dir = lab.parentElement!;
        if (dir.classList.contains("dir-space")) {
            this.checkAndSet(dir.id.slice(4));
        }
    }

    renderFiles(files: any[]) {
        let items: JSX.Element[] = []
        for (const file of files) {
            if (file.Child == null) {
                items.push(<label className="checkbox-file">{file.Name} <span className="size">[{file.Size}]</span>
                    <input id={"file_"+file.Path} type="checkbox" className="file-to-download" defaultChecked={true} onInput={(e) => {
                        let dir = e.currentTarget.parentElement!.parentElement!;
                        if (dir.classList.contains("dir-space")) {
                            this.checkAndSet(dir.id.slice(4))
                        }
                    }}/>
                    <span className="checkmark"></span>
                </label>)
            } else {
                let id = "dir_"+file.Path;
                let idLabel = "lab_"+file.Path;
                items.push(<label id={idLabel} className="checkbox-file folder">{file.Name} <span className="size">[{file.Size}]</span>
                    <input type="checkbox" defaultChecked={true} onInput={(e)=> {
                        console.log(e.target);
                        let dep = document.getElementById(id)!;
                        for (const el of dep.getElementsByClassName("checkbox-file")) {
                            (el.children.item(1) as HTMLInputElement)!.checked = e.currentTarget.checked
                        }

                        let dir = e.currentTarget.parentElement!.parentElement!;
                        if (dir.classList.contains("dir-space")) {
                            this.checkAndSet(dir.id.slice(4))
                        }
                    }}/>
                    <span className="checkmark"></span>
                    <button onClick={(e) => {
                        let dep = document.getElementById(id)!;
                        if (dep.style.display == "none") {
                            dep.style.display = "flex";
                            e.currentTarget.style.transform = "rotate(180deg)";
                        } else {
                            dep.style.display = "none";
                            e.currentTarget.style.transform = "";
                        }
                    }}/>
                </label>)
                items.push(<div style={{display:"none"}} className="dir-space" id={id}>{this.renderFiles(file.Child)}</div>)
            }
        }
        return items;
    }

    render() {
        return <Modal allowClose={true} onHide={this.cancel} content={(
            <>
                <div style={(this.state.selectFilesStage && this.state.files.length == 0) ? {} : {display: "none"}} className="add-torrent-block">
                    <span className="title">Searching for torrent info...</span>
                    <div className="files-selector">
                        <div className="loader-block"><span className="loader"/></div>
                    </div>
                </div>
                <div style={(this.state.selectFilesStage && this.state.files.length > 0) ? {} : {display: "none"}} className="add-torrent-block">
                    <span className="title" style={{marginBottom: "5px"}}>Select files</span>
                    <div className="files-selector">
                        {this.renderFiles(this.state.files)}
                    </div>
                </div>
                <div style={this.state.selectFilesStage ? {display: "none"} : {width: "287px"}} className="add-torrent-block">
                    <span className="title">Add Torrent</span>
                    <input id="torrent-hash-field" required={true} autoFocus={true} placeholder="Insert Bag ID..." onChange={(v) => {
                        this.setState((current) => ({...current, err: this.state.err, fieldMeta: undefined, fieldHash: v.target.value,
                            canContinue: v.target.value.length == 64}));
                        (document.getElementById("file-select") as HTMLInputElement).value = "";
                    }} value={this.state.fieldHash} type="text"/>
                    <hr className="hr-text" data-content="or"/>
                    <div className={"file-selector-ui "+ (this.state.fieldMeta !== undefined ? "selected" : "" )}>
                        <img src={this.state.fieldMeta !== undefined ? (this.props.isDark ? FileDark : FileLight) : Upload}/>
                        <label className="big">{this.state.fieldMeta !== undefined ? "File added" : "Select .tonbag file" }</label>
                        <label>{this.state.fieldMeta !== undefined ? this.state.fileName : "Click or drag and drop file here."}</label>
                    </div>
                    <input id="file-select" type="file" className="file" accept=".tonbag" required={true} onInput={(e)=> {
                        let reader = new FileReader();
                        let fileInput = e.target as HTMLInputElement;
                        if (fileInput && fileInput.files) {
                            let name = fileInput.files[0].name;
                            if(!name.endsWith(".tonbag")) {
                                return;
                            }

                            reader.readAsArrayBuffer(fileInput.files[0]);
                            reader.onload = (ev) => {
                                if (ev.type === "load") {
                                    this.setState((current) => ({...current, fileName: name, fieldHash: undefined, fieldMeta: reader.result as ArrayBuffer, canContinue: true}))
                                }
                            }
                        }
                    }}/>
                    <span className="error">{this.state.err}</span>
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.cancel}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={!this.state.canContinue} onClick={()=>{this.next()}}>
                        Continue
                    </button>
                </div>
            </>
        )}/>;
    }
}
