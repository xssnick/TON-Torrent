import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetInfo, RemoveTorrent} from "../../wailsjs/go/main/App";
import {Refresh} from "./Table";
import FileLight from "../../public/light/file-popup.svg";
import FileDark from "../../public/dark/file-popup.svg";
import {EventsEmit} from "../../wailsjs/runtime";

interface ReinitTunnelConfirmProps {
    onExit: () => void
}

export class ReinitTunnelConfirm extends Component<ReinitTunnelConfirmProps, {}> {
    constructor(props: any) {
        super(props);
    }

    next = async (agree: boolean) => {
        EventsEmit("tunnel_reinit_ask_result", agree);
        this.props.onExit();
    }

    render() {
        return baseModal(this.props.onExit, (
            <>
                <div className="add-torrent-block">
                    <span className="title">Tunnel seems stalled, do you want to reinit it?</span>
                    <span className="title">Keep in mind that new payment channels could be opened if tunnel is not free.</span>
                </div>
                <div className="modal-control">
                    <button className="second-button" style={{width:"115px"}} onClick={()=>{this.next(false)}}>
                        No, just wait
                    </button>
                    <button className="main-button" style={{width:"115px"}} onClick={()=>{this.next(true)}}>
                        Yes, reinit
                    </button>
                </div>
            </>
        ));
    }
}