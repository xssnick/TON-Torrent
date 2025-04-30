import React from 'react';
import ReactDOM from 'react-dom';

export const baseModal = (onHide: () => void, content: JSX.Element, depth: number = 0) => {
    const modal = (
        <>
            <div className="modal-backdrop" style={{zIndex: 1000+depth}} onClick={onHide}/>
            <div className="modal-wrap" style={{zIndex: 1000+depth+1}}>
                <div className="modal-back" style={{zIndex: 1000+depth+1}}>
                    <div className="modal-content">{content}</div>
                </div>
            </div>
        </>
    );
    return ReactDOM.createPortal(modal, document.body);
}