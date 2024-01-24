import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {
    CancelCreateTorrent,
    CreateTorrent,
    ExportMeta,
    FetchProviderRates,
    OpenDir,
    OpenFolderSelectFile
} from "../../wailsjs/go/main/App";
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime";

interface State {
    fetchStage: boolean
    canContinue: boolean

    providerKey: string
    err?: string
}

interface AddProviderModalProps {
    hash: string
    onExit: () => void;
}

export class AddProviderModal extends Component<AddProviderModalProps, State> {
    constructor(props: AddProviderModalProps) {
        super(props);
        this.state = {
            err: "",
            fetchStage: false,
            canContinue: false,
            providerKey: "",
        }
    }

    componentDidMount() {
        EventsOn("provider-connected", () => {
            this.props.onExit()
        })
    }
    componentWillUnmount() {
        EventsOff("provider-connected")
    }

    next = () => {
        this.setState((current) => ({ ...current, canContinue: false, fetchStage: true }));

        console.log(this.props.hash+" -- "+ this.state.providerKey)
        FetchProviderRates(this.props.hash, this.state.providerKey).then((rates) => {
            console.log(rates);

            if (rates.Success) {
                EventsEmit("provider-added", rates.Provider, this.props.hash);
                this.props.onExit();
            } else {
                this.setState((current) => ({...current, err: rates.Reason, canContinue: true, fetchStage: false}));
            }
        });
    }

    render() {
        return baseModal(this.props.onExit, (
            <>
                <div style={this.state.fetchStage ? {width: "287px"} : {display: "none"}} className="add-torrent-block">
                    <span className="title">Connecting to provider...</span>
                    <div className="files-selector">
                        <div className="loader-block"><span className="loader"/><span className="loader-text"></span></div>
                    </div>
                </div>
                <div style={this.state.fetchStage ? {display: "none"} : {width: "287px"}} className="add-torrent-block">
                    <span className="title">Add Provider</span>
                    <input className="torrent-name-input" placeholder={"Key"} maxLength={100} autoFocus={true} onInput={(e) => {
                        let val = e.currentTarget.value;
                        let can = Boolean(val.length == 64 && val.match(/^[0-9a-f]+$/i));
                        this.setState((current) => ({...current, name: val, canContinue: can, providerKey: can ? val : ""}));
                    }}/>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                {(this.state.fetchStage) ? <div className="modal-control">
                        <button className="second-button" style={{width: "100%"}} onClick={this.props.onExit}>
                            Cancel
                        </button>
                    </div>:
                    <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={!this.state.canContinue} onClick={()=>{this.next()}}>
                        Continue
                    </button>
                </div>}
            </>
        ));
    }
}