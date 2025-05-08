import React from 'react';
import ReactDOM from 'react-dom';

export const baseModal = (onHide: () => void, content: JSX.Element) => {
    const modal = (
        <>
            <div className="modal-backdrop"/>
            <div className="modal-wrap">
                <div className="modal-back">
                    <div className="modal-content">{content}</div>
                </div>
            </div>
        </>
    );
    return ReactDOM.createPortal(modal, document.body);
}