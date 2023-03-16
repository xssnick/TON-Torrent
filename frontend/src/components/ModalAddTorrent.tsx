import React, {Component} from 'react';
import {AddTorrentByHash, CheckHeader, GetFiles, RemoveTorrent} from "../../wailsjs/go/main/App";
import Expand from "../assets/images/icons/expand.svg";
import {EventsOn} from "../../wailsjs/runtime";
import {refresh} from "./Table";
import {baseModal} from "./Modal";

interface State {
    selectFilesStage: boolean
    field: string
    err: string

    hash?: string
    files: any[]
}

interface AddTorrentModalProps {
    onExit: () => void;
}

export class AddTorrentModal extends Component<AddTorrentModalProps, State> {
    inter?: number

    cancel = () => {
        if (this.state.hash) {
            RemoveTorrent(this.state.hash, true).then(refresh)
        }
        this.props.onExit()
    }

    next = () => {
        if (!this.state.selectFilesStage) {
            let hash = this.state.field;
            AddTorrentByHash(hash).then((err) => {
                if (err == "") {
                    this.setState({err: "", field: "", selectFilesStage: true, hash})
                    return
                }
                this.setState({err: err, field: ""})
            }).then(() => {
                // check header availability every 100 ms, and load files list when we get it
                this.inter = setInterval(() => {
                    CheckHeader(hash).then(has => {
                        if (has) {
                            GetFiles(hash).then((tree)=> {
                                this.setState({
                                    selectFilesStage: this.state.selectFilesStage,
                                    field: this.state.field,
                                    err: this.state.err,
                                    hash: this.state.hash,
                                    files: tree
                                });
                            }).then(() => { clearInterval(this.inter) })
                        }
                    })
                }, 100)
            })
        } else {
            this.props.onExit()
        }
    }

    componentWillUnmount() {
        if (this.inter)
            clearInterval(this.inter)
    }

    componentWillMount() {
        this.setState({
            selectFilesStage: false,
            field: "",
            err: "",
            files: [],
        });
    }

    renderFiles(files: any[]) {
        let items: JSX.Element[] = []
        for (const file of files) {
            if (file.Child == null) {
                items.push(<label className="checkbox-file">{file.Name} <span className="size">[{file.Size}]</span>
                    <input type="checkbox" defaultChecked={true}/>
                    <span className="checkmark"></span>
                </label>)
            } else {
                let id = file.Path.replace('/','_');
                items.push(<label className="checkbox-file folder">{file.Name} <span className="size">[{file.Size}]</span>
                    <input type="checkbox" defaultChecked={true} onInput={(e)=> {
                        let dep = document.getElementById(id)!;
                        for (const el of dep.getElementsByClassName("checkbox-file")) {
                            (el.children.item(1) as HTMLInputElement)!.checked = e.currentTarget.checked
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
        return baseModal(this.cancel, (
            <>
                <div style={this.state.selectFilesStage ? {} : {display: "none"}} className="add-torrent-block">
                    <div className="files-selector">
                        {this.state.files.length > 0 ? this.renderFiles(this.state.files) : <label>Loading...</label>}
                    </div>
                </div>
                <div style={this.state.selectFilesStage ? {display: "none"} : {}} className="add-torrent-block">
                    <span className="title">Bag ID</span>
                    <input id="torrent-hash-field" required={true} placeholder="0000000000000000000000000000000000000000000000000000000000000000" onChange={(v) => {
                        this.setState({
                            field: v.target.value,
                            err: this.state.err,
                        });

                        // if (v.target.value.length == 64)
                        //    this.add(v.target.value);
                    }} value={this.state.field} type="text"/>
                    <span className="error">{this.state.err}</span>
                    <hr className="hr-text" data-content="OR"/>
                    <input type="file" className="file" accept=".tontorrent" required={true} onInput={(e)=> {
                        let reader = new FileReader();
                        let fileInput = e.target as HTMLInputElement;
                        if (fileInput && fileInput.files) {
                            reader.readAsBinaryString(fileInput.files[0]);
                            reader.onload = (ev) => {
                                if (ev.type === "load") {
                                    console.log(reader.result);
                                }
                            }
                        }
                    }}/>
                </div>
                <div className="modal-control">
                    <button className="item" onClick={this.cancel}>
                        Cancel
                    </button>
                    <button className="item main" onClick={()=>{this.next()}}>
                        Continue
                    </button>
                </div>
            </>
        ));
    }
}