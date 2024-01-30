import React, { useState, useEffect } from 'react';
import { EventsOff, EventsOn } from "../../wailsjs/runtime";
import { GetInfo } from "../../wailsjs/go/main/App";
import { textState } from "./Table";

export interface InfoProps {
    torrent: string;
}

interface State {
    left: string;
    path: string;
    description: string;
    downloadSpeed: string;
    uploadSpeed: string;
    downloaded: string;
    size: string;
    status: string;
    added: string;
    peers: string;
    progress: string;
    uploaded: string;
    ratio: string;
}

 const InfoTorrentMenu: React.FC<InfoProps> = (props) => {
    const [state, setState] = useState<State>({
        left: "",
        path: "",
        description: "",
        downloadSpeed: "",
        uploadSpeed: "",
        downloaded: "",
        size: "",
        status: "",
        added: "",
        peers: "",
        progress: "",
        uploaded: "",
        ratio: "",
    });

    const short = (val: string) => {
        return val; //.length > 45 ? val.slice(0, 45)+ "..." : val
    };

    const copy = (text: string) => {
        return (clicked: any) => {
            navigator.clipboard.writeText(text).then();
            clicked.target.classList.add('clicked');
            setTimeout(() => {
                clicked.target.classList.remove('clicked');
            }, 500);
        };
    };

    const update = () => {
        GetInfo(props.torrent).then((tr: any) => {
            setState({
                left: tr.TimeLeft,
                path: tr.Path,
                progress: tr.Progress,
                description: tr.Description,
                downloadSpeed: tr.Download,
                uploadSpeed: tr.Upload,
                downloaded: tr.Downloaded,
                size: tr.Size,
                status: tr.State,
                added: tr.AddedAt,
                peers: tr.Peers,
                uploaded: tr.Uploaded,
                ratio: tr.Ratio,
            });
        });
    };

    useEffect(() => {
        update();
        EventsOn("update_info", () => {
            update();
        });

        return () => {
            EventsOff("update_info");
        };
    }, [props.torrent]);

    return (
        <div>
            <div className="info-data-block">
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Progress</span></div>
                    <div className="item" style={{ flexGrow: "1" }}>
                        <span className="value" style={{ width: "43px" }}>{state.progress}%</span>
                        <div className="info-progress-bar">
                            <div className="filled" style={{ width: state.progress + "%" }} />
                        </div>
                    </div>
                </div>
                {(state.status === "downloading" || state.status === "seeding") ? <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Upload Speed</span></div>
                    <div className="item" style={{ width: "20%" }}><span className="value">{state.uploadSpeed}</span></div>
                    {state.status === "downloading" ? <>
                        <div className="item" style={{ width: "20%" }}><span className="field">Download Speed</span></div>
                        <div className="item" style={{ width: "15%" }}><span className="value">{state.downloadSpeed}</span></div>
                        <div className="item" style={{ width: "12%" }}><span className="field">Remaining</span></div>
                        <div className="item" style={{ flexGrow: "1", justifyContent: "flex-end" }}><span className="value">{state.left}</span></div>
                    </> : <>
                        <div className="item" style={{ width: "20%" }}><span className="field">Uploaded</span></div>
                        <div className="item" style={{ width: "15%" }}><span className="value">{state.uploaded}</span></div>
                        <div className="item" style={{ width: "12%" }}><span className="field">Ratio</span></div>
                        <div className="item" style={{ flexGrow: "1", justifyContent: "flex-end" }}><span className="value">{state.ratio}</span></div>
                    </>}
                </div> : <></>}
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Downloaded</span></div>
                    <div className="item" style={{ width: "20%" }}><span className="value">{state.downloaded}</span></div>
                    <div className="item" style={{ width: "20%" }}><span className="field">Peers</span></div>
                    <div className="item" style={{ width: "15%" }}><span className="value">{state.peers}</span></div>
                    <div className="item" style={{ width: "12%" }}><span className="field">Size</span></div>
                    <div className="item" style={{ flexGrow: "1", justifyContent: "flex-end" }}><span className="value">{state.size}</span></div>
                </div>
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Path</span></div>
                    <div className="item" style={{ flexGrow: "1", maxWidth: "75%" }}>
                        <span className="value" style={{ maxWidth: "90%" }}>{short(state.path)}</span>
                        <button onClick={copy(state.path)} />
                    </div>
                </div>
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Bag ID</span></div>
                    <div className="item" style={{ flexGrow: "1", maxWidth: "75%" }}>
                        <span className="value" style={{ maxWidth: "90%" }}>{short(props.torrent.toUpperCase())}</span>
                        <button onClick={copy(props.torrent)} />
                    </div>
                </div>
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Added at</span></div>
                    <div className="item" style={{ flexGrow: "1" }}><span className="value">{state.added}</span></div>
                </div>
                <div className="basic">
                    <div className="item" style={{ width: "20%" }}><span className="field">Status</span></div>
                    <div className="item" style={{ flexGrow: "1" }}><span className="value">{textState(state.status, Number(state.peers))}</span></div>
                </div>
                {state.description !== "" ?
                    <div className="basic">
                        <div className="item" style={{ width: "20%" }}><span className="field">Name</span></div>
                        <div className="item" style={{ flexGrow: "1", maxWidth: "80%" }}><span className="value">{short(state.description)}</span></div>
                    </div> : ""}
            </div>
        </div>
    );
};

export default InfoTorrentMenu;
