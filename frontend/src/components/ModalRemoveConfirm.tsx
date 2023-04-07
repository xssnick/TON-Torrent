import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetInfo, RemoveTorrent} from "../../wailsjs/go/main/App";
import {Refresh} from "./Table";

interface State {
    name: string
}

interface RemoveConfirmModalProps {
    hash: string
    onExit: () => void;
}

export class RemoveConfirmModal extends Component<RemoveConfirmModalProps, State> {
    constructor(props: any) {
        super(props);

        this.state = {
            name: "",
        };
    }

    componentDidMount() {
        GetInfo(this.props.hash).then((info:any)=>{
            this.setState((current)=>({...current, name: info.Description}))
        })
    }

    next = (removeFiles: boolean) => {
        RemoveTorrent(this.props.hash, removeFiles, false).then(Refresh)
        this.props.onExit()
    }

    render() {
        return baseModal(this.props.onExit, (
            <>
                <div className="add-torrent-block">
                    <span className="title big">Are you sure you want to delete torrent?</span>
                    <span className="title name">"{this.state.name}"</span>
                </div>
                <div className="modal-control" style={{width:"400px"}}>
                    <button className="second-button" style={{width:"115px"}} onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="second-button danger" style={{width:"115px"}}  onClick={()=>{this.next(true)}}>
                        Yes, delete all
                    </button>
                    <button className="main-button" style={{width:"115px"}} onClick={()=>{this.next(false)}}>
                        Yes, keep files
                    </button>
                </div>
            </>
        ));
    }
}