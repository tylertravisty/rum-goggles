import { useEffect, useRef, useState } from 'react';

import { ChevronDown } from '../assets';
import './DropDown.css';

export function DropDown(props) {
    const [options, setOptions] = useState(props.options !== undefined ? props.options : []);
    const [selected, setSelected] = useState(props.selected !== undefined ? props.selected : '');
    const [toggled, setToggled] = useState(false);
    const toggle = () => {
        setToggled(!toggled);
    };

    useEffect(() => {
        setSelected(props.selected !== undefined ? props.selected : '');
    }, [props.selected]);

    useEffect(() => {
        setOptions(props.options !== undefined ? props.options : []);
    }, [props.options]);

    const select = (option) => {
        props.select(option);
        setSelected(option);
        toggle();
    };

    return (
        <div className='dropdown'>
            <button className='dropdown-toggle' onClick={toggle}>
                <div style={{ width: '20px' }}></div>
                <span className='dropdown-toggle-text'>{selected}</span>
                <img className='dropdown-toggle-icon' src={ChevronDown} />
            </button>
            {toggled && (
                <DropDownMenu
                    options={options}
                    select={select}
                    selected={selected}
                    toggle={toggle}
                />
            )}
        </div>
    );
}

function DropDownMenu(props) {
    const menuRef = useRef();
    const { width } = menuWidth(menuRef);

    return (
        <div className='dropdown-menu-container' ref={menuRef}>
            {width !== undefined && (
                <div className='dropdown-menu' style={{ width: width + 'px' }}>
                    {props.options.map((option, index) => (
                        <button
                            className={
                                props.selected === option
                                    ? 'dropdown-menu-option dropdown-menu-option-selected'
                                    : 'dropdown-menu-option'
                            }
                            key={index}
                            onClick={() => props.select(option)}
                        >
                            {option}
                        </button>
                    ))}
                </div>
            )}
            <div className='dropdown-menu-background' onClick={props.toggle}></div>
        </div>
    );
}

export const menuWidth = (menuRef) => {
    const [width, setWidth] = useState(0);

    useEffect(() => {
        const getWidth = () => ({ width: menuRef.current.offsetWidth });

        const handleResize = () => {
            setWidth(getWidth());
        };

        if (menuRef.current) {
            setWidth(getWidth());
        }

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
        };
    }, [menuRef]);

    return width;
};
