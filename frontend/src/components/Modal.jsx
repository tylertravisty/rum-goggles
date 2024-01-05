import { XLg } from '../assets/icons';
import './Modal.css';

export function Modal(props) {
    return (
        <div
            className='modal-background'
            onClick={props.onClose}
            style={{ zIndex: props.show ? 10 : -10 }}
        >
            <div
                className='modal-container'
                onClick={(event) => event.stopPropagation()}
                style={props.style}
            >
                <div className='modal-header'>
                    <span className='modal-title'>{props.title}</span>
                    <button className='modal-close' onClick={props.onClose}>
                        <img className='modal-close-icon' src={XLg} />
                    </button>
                </div>
                <div className='modal-body'>{props.children}</div>
                <div className='modal-footer'>
                    {props.cancelButton && (
                        <button className='modal-button-cancel' onClick={props.onCancel}>
                            {props.cancelButton}
                        </button>
                    )}
                    {props.deleteButton && (
                        <button className='modal-button-delete' onClick={props.onDelete}>
                            {props.deleteButton}
                        </button>
                    )}
                    {props.submitButton && (
                        <button className='modal-button' onClick={props.onSubmit}>
                            {props.submitButton}
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
}

export function SmallModal(props) {
    return (
        <div
            className='modal-background'
            onClick={props.onClose}
            style={{ zIndex: props.show ? 10 : -10 }}
        >
            <div
                className='small-modal-container'
                onClick={(event) => event.stopPropagation()}
                style={props.style}
            >
                <div className='small-modal-header'>
                    <span className='small-modal-title'>{props.title}</span>
                    <button className='modal-close' onClick={props.onClose}>
                        <img className='modal-close-icon' src={XLg} />
                    </button>
                </div>
                <div className='modal-body'>
                    <span className='small-modal-message'>{props.message}</span>
                </div>
                <div className='small-modal-footer'>
                    {props.cancelButton && (
                        <button className='modal-button-cancel' onClick={props.onCancel}>
                            {props.cancelButton}
                        </button>
                    )}
                    {props.deleteButton && (
                        <button className='small-modal-button-delete' onClick={props.onDelete}>
                            {props.deleteButton}
                        </button>
                    )}
                    {props.submitButton && (
                        <button className='modal-button' onClick={props.onSubmit}>
                            {props.submitButton}
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
}
