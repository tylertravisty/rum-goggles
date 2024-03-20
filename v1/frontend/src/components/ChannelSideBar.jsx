import { useEffect, useState } from 'react';
import { Modal, SmallModal } from './Modal';
import { AccountList, AddChannel, Login } from '../../wailsjs/go/main/App';

import { ChevronRight, CircleGreenBackground, Eye, EyeSlash, PlusCircle } from '../assets';
import './ChannelSideBar.css';

function ChannelSideBar(props) {
    const [accounts, setAccounts] = useState({});
    const [error, setError] = useState('');
    const [addOpen, setAddOpen] = useState(false);
    const [refresh, setRefresh] = useState(false);
    const [scrollY, setScrollY] = useState(0);

    useEffect(() => {
        AccountList()
            .then((response) => {
                setAccounts(response);
            })
            .catch((error) => {
                setError(error);
            });
    }, [refresh]);

    const sortAccounts = () => {
        let keys = Object.keys(accounts);

        let sorted = [...keys].sort((a, b) =>
            accounts[a].account.username.toLowerCase() > accounts[b].account.username.toLowerCase()
                ? 1
                : -1
        );

        return sorted;
    };

    const handleScroll = (event) => {
        setScrollY(event.target.scrollTop);
    };

    return (
        <>
            <ModalAdd
                onClose={() => setAddOpen(false)}
                onRefresh={() => {
                    setRefresh(!refresh);
                }}
                show={addOpen}
            />
            <div className='channel-sidebar'>
                <div className='channel-sidebar-body' onScroll={handleScroll}>
                    {sortAccounts().map((account, index) => (
                        <AccountChannels
                            account={accounts[account]}
                            key={index}
                            scrollY={scrollY}
                            top={index === 0}
                        />
                    ))}
                </div>
                <div className='channel-sidebar-footer'>
                    <ButtonIcon
                        hoverText={'Add an account/channel'}
                        onClick={() => setAddOpen(true)}
                        scrollY={0}
                    />
                </div>
            </div>
        </>
    );
}

export default ChannelSideBar;

function AccountChannels(props) {
    const sortChannels = () => {
        let sorted = [...props.account.channels].sort((a, b) =>
            a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1
        );

        return sorted;
    };

    if (props.account.account !== undefined) {
        return (
            <div
                className='channel-sidebar-account-list'
                style={props.top ? { borderTop: 'none' } : {}}
            >
                <AccountIcon account={props.account.account} key={0} scrollY={props.scrollY} />
                {sortChannels().map((channel, index) => (
                    <ChannelIcon channel={channel} key={index + 1} scrollY={props.scrollY} />
                ))}
            </div>
        );
    }
}

function AccountIcon(props) {
    const [hover, setHover] = useState(false);

    return (
        <div
            className='channel-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            {props.account.profile_image === null ? (
                <span className='channel-sidebar-icon-initial'>
                    {props.account.username[0].toUpperCase()}
                </span>
            ) : (
                <img className='channel-sidebar-icon-image' src={props.account.profile_image} />
            )}
            <img className='channel-sidebar-icon-account' src={CircleGreenBackground} />
            {hover && (
                <HoverName name={'/user/' + props.account.username} scrollY={props.scrollY} />
            )}
        </div>
    );
}

function ButtonIcon(props) {
    const [hover, setHover] = useState(false);

    return (
        <div
            className='channel-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            <button className='channel-sidebar-button' onClick={props.onClick}>
                <img className='channel-sidebar-button-icon' src={PlusCircle} />
            </button>
            {hover && <HoverName name={props.hoverText} scrollY={props.scrollY} />}
        </div>
    );
}

function ChannelIcon(props) {
    const [hover, setHover] = useState(false);
    return (
        <div
            className='channel-sidebar-icon'
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
        >
            {props.channel.profile_image === null ? (
                <span className='channel-sidebar-icon-initial'>
                    {props.channel.name[0].toUpperCase()}
                </span>
            ) : (
                <img className='channel-sidebar-icon-image' src={props.channel.profile_image} />
            )}
            {hover && (
                <HoverName
                    name={'/c/' + props.channel.name.replace(/\s/g, '')}
                    scrollY={props.scrollY}
                />
            )}
        </div>
    );
}

function HoverName(props) {
    return (
        <div
            className='channel-sidebar-icon-hover'
            style={{ transform: 'translate(75px, -' + (50 + props.scrollY) + 'px)' }}
        >
            <span className='channel-sidebar-icon-hover-text'>{props.name}</span>
        </div>
    );
}

function ModalAdd(props) {
    const [accountPassword, setAccountPassword] = useState('');
    const [accountPasswordValid, setAccountPasswordValid] = useState(true);
    const updateAccountPassword = (event) => {
        if (loading()) {
            return;
        }
        setAccountPassword(event.target.value);
    };
    const [accountUsername, setAccountUsername] = useState('');
    const [accountUsernameValid, setAccountUsernameValid] = useState(true);
    const updateAccountUsername = (event) => {
        if (loading()) {
            return;
        }
        setAccountUsername(event.target.value);
    };
    const [addAccountLoading, setAddAccountLoading] = useState(false);
    const [addChannelLoading, setAddChannelLoading] = useState(false);
    const [channelKey, setChannelKey] = useState('');
    const [channelKeyValid, setChannelKeyValid] = useState(true);
    const updateChannelKey = (event) => {
        if (loading()) {
            return;
        }
        setChannelKey(event.target.value);
    };
    const [error, setError] = useState('');
    const [stage, setStage] = useState('start');

    useEffect(() => {
        if (addAccountLoading) {
            Login(accountUsername, accountPassword)
                .then(() => {
                    reset();
                    props.onClose();
                    props.onRefresh();
                })
                .catch((error) => {
                    setAddAccountLoading(false);
                    setError(error);
                });
        }
    }, [addAccountLoading]);

    useEffect(() => {
        if (addChannelLoading) {
            AddChannel(channelKey)
                .then(() => {
                    reset();
                    props.onClose();
                    props.onRefresh();
                })
                .catch((error) => {
                    setAddChannelLoading(false);
                    setError(error);
                });
        }
    }, [addChannelLoading]);

    const back = () => {
        if (loading()) {
            return;
        }
        reset();
    };

    const close = () => {
        if (loading()) {
            return;
        }
        reset();
        props.onClose();
    };

    const reset = () => {
        setStage('start');
        resetAccount();
        resetChannel();
    };

    const add = () => {
        switch (stage) {
            case 'account':
                addAccount();
                break;
            case 'channel':
                addChannel();
                break;
            default:
                close();
        }
    };

    const addAccount = () => {
        if (loading()) {
            return;
        }

        if (accountUsername === '') {
            setAccountUsernameValid(false);
            return;
        }

        if (accountPassword === '') {
            setAccountPasswordValid(false);
            return;
        }

        setAddAccountLoading(true);
    };

    const addChannel = () => {
        if (loading()) {
            return;
        }

        if (channelKey === '') {
            setChannelKeyValid(false);
            return;
        }

        setAddChannelLoading(true);
    };

    const loading = () => {
        return addAccountLoading || addChannelLoading;
    };

    const resetAccount = () => {
        setAccountPassword('');
        setAccountPasswordValid(true);
        setAccountUsername('');
        setAccountUsernameValid(true);
        setAddAccountLoading(false);
    };

    const resetChannel = () => {
        setChannelKey('');
        setChannelKeyValid(true);
        setAddChannelLoading(false);
    };

    return (
        <>
            <SmallModal
                onClose={() => setError('')}
                show={error !== ''}
                style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                title={'Error'}
                message={error}
                submitButton={'OK'}
                onSubmit={() => setError('')}
            />
            <Modal
                cancelButton={stage !== 'start' ? 'Back' : ''}
                onCancel={back}
                onClose={close}
                show={props.show}
                style={{ height: '480px', minHeight: '480px', width: '360px', minWidth: '360px' }}
                submitButton={stage !== 'start' ? 'Add' : ''}
                submitLoading={loading()}
                onSubmit={add}
            >
                {stage === 'start' && <ModalAddStart setStage={setStage} />}
                {stage === 'account' && (
                    <ModalAddAccount
                        accountPassword={accountPassword}
                        accountPasswordValid={accountPasswordValid}
                        updateAccountPassword={updateAccountPassword}
                        accountUsername={accountUsername}
                        accountUsernameValid={accountUsernameValid}
                        updateAccountUsername={updateAccountUsername}
                    />
                )}
                {stage === 'channel' && (
                    <ModalAddChannel
                        channelKey={channelKey}
                        channelKeyValid={channelKeyValid}
                        updateChannelKey={updateChannelKey}
                    />
                )}
            </Modal>
        </>
    );
}

function ModalAddAccount(props) {
    const [showKey, setShowKey] = useState(false);
    const updateShowKey = () => setShowKey(!showKey);
    const [showPassword, setShowPassword] = useState(false);
    const updateShowPassword = () => setShowPassword(!showPassword);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Add Account</span>
                <span className='modal-add-account-channel-subtitle'>
                    Log into your Rumble account
                </span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.accountUsernameValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        USERNAME - Please enter a valid username
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>USERNAME</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-text'
                        onChange={!props.loading && props.updateAccountUsername}
                        placeholder={'Username'}
                        type={'text'}
                        value={props.accountUsername}
                    ></input>
                </div>
                {props.accountPasswordValid === false ? (
                    <label className='modal-add-account-channel-label-warning'>
                        PASSWORD - Please enter a valid password
                    </label>
                ) : (
                    <label className='modal-add-account-channel-label'>PASSWORD</label>
                )}
                <div className='modal-add-account-channel-input'>
                    <input
                        className='modal-add-account-channel-input-password'
                        onChange={!props.loading && props.updateAccountPassword}
                        placeholder={'Password'}
                        type={showPassword ? 'text' : 'password'}
                        value={props.accountPassword}
                    ></input>
                    <button
                        className='modal-add-account-channel-input-show'
                        onClick={updateShowPassword}
                    >
                        <img
                            className='modal-add-account-channel-input-show-icon'
                            src={showPassword ? EyeSlash : Eye}
                        />
                    </button>
                </div>
            </div>
            <div></div>
        </div>
    );
}

function ModalAddChannel(props) {
    const [showKey, setShowKey] = useState(false);
    const updateShowKey = () => setShowKey(!showKey);

    return (
        <div className='modal-add-account-channel'>
            <div className='modal-add-account-channel-header'>
                <span className='modal-add-account-channel-title'>Add Channel</span>
                <span className='modal-add-account-channel-subtitle'>
                    Copy an API key below to add a channel
                </span>
            </div>
            <div className='modal-add-account-channel-body'>
                {props.channelKeyValid === false ? (
                    <label className='modal-add-channel-label-warning'>
                        API KEY - Please enter a valid API key
                    </label>
                ) : (
                    <label className='modal-add-channel-label'>API KEY</label>
                )}
                <div className='modal-add-channel-key'>
                    <input
                        className='modal-add-channel-key-input'
                        onChange={!props.loading && props.updateChannelKey}
                        placeholder={'Enter API key'}
                        type={showKey ? 'text' : 'password'}
                        value={props.channelKey}
                    ></input>
                    <button className='modal-add-channel-key-show' onClick={updateShowKey}>
                        <img
                            className='modal-add-channel-key-show-icon'
                            src={showKey ? EyeSlash : Eye}
                        />
                    </button>
                </div>
                <span className='modal-add-channel-description'>API KEYS SHOULD LOOK LIKE</span>
                <span className='modal-add-channel-description-subtext'>
                    https://rumble.com/-livestream-api/get-data?key=really-long_string-of_random-characters
                </span>
            </div>
            <div></div>
        </div>
    );
}

function ModalAddStart(props) {
    return (
        <div className='modal-add-account-channel'>
            <span className='modal-add-account-channel-title'>Add an Account or Channel</span>
            <div className='modal-add-account-channel-body'>
                <button
                    className='modal-add-account-channel-button'
                    onClick={() => props.setStage('account')}
                >
                    <div className='modal-add-account-channel-button-left'>
                        <span>Add Account</span>
                    </div>
                    <img
                        className='modal-add-account-channel-button-right-icon'
                        src={ChevronRight}
                    />
                </button>
                <button
                    className='modal-add-account-channel-button'
                    onClick={() => props.setStage('channel')}
                >
                    <div className='modal-add-account-channel-button-left'>
                        <span>Add Channel</span>
                    </div>
                    <img
                        className='modal-add-account-channel-button-right-icon'
                        src={ChevronRight}
                    />
                </button>
            </div>
            <div></div>
        </div>
    );
}
