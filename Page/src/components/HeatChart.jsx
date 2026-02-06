import React, {useEffect, useMemo, useState} from 'react';
import {Avatar, Modal, ModalBody, ModalHeader, Select, SelectItem, Tooltip} from "@heroui/react";
import axios from "axios";


const getColor = (count) => {
    if (count === 0) return '#ebedf0';
    if (count < 3) return '#c6e48b';
    if (count < 6) return '#7bc96f';
    if (count < 9) return '#239a3b';
    return '#196127';
};
const HeatMap = (props) => {
    return (
        <div style={{ display: 'flex' }}>

            <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between', marginRight: '10px' }}>
                {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(d => <span key={d} style={{fontSize: '12px', height: '20px'}}>{d}</span>)}
            </div>

            <div
                style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(24, 1fr)',
                    gridTemplateRows: 'repeat(7, 1fr)',
                    gap: '4px'
                }}
            >
                {props.data.map((d, index) => (

                    <Tooltip key={index} content={ !!(
                        d.Message || d.Money
                        ) &&
                        <div>
                            <p>Message: {d.Message }</p>
                            <p>Money: {d.Money }</p>
                            <div className={'flex flex-row mt-4'}>
                                {d.Livers.map((item) => {
                                    return (
                                        <div className={'flex flex-col items-center'}>
                                            <Avatar src={`${AVATAR_API}${item.UID}`}></Avatar>
                                            {item.UName}
                                        </div>

                                    )
                                })}
                            </div>
                        </div>
                    } placement="right" >
                        <div
                            className={'w-[12px] h-[12px]'}
                            style={
                                {
                                    backgroundColor: getColor(d.Message??d.Money),
                                }
                            }
                            key={index}
                        />
                    </Tooltip>

                ))}
            </div>
        </div>
    );
}
export default HeatMap;

const getWeekDateRange = (year, week) => {
    const simple = new Date(year, 0, 4);
    const day = simple.getDay() || 7;
    const mondayOfWeek1 = new Date(simple.getFullYear(), 0, 4 - day + 1);
    const startObj = new Date(mondayOfWeek1);
    startObj.setDate(mondayOfWeek1.getDate() + (week - 1) * 7);
    const endObj = new Date(startObj);
    endObj.setDate(startObj.getDate() + 6);

    const format = (date) => {
        const m = (date.getMonth() + 1).toString().padStart(2, '0');
        const d = date.getDate().toString().padStart(2, '0');
        return `${m}月${d}日`;
    };
    return `${format(startObj)} - ${format(endObj)}`;
};
const getISOWeekNumber = (date) => {
    const target = new Date(date.valueOf());
    const dayNumber = (date.getDay() + 6) % 7
    target.setDate(target.getDate() - dayNumber + 3)
    const firstThursday = target.valueOf();

    target.setMonth(0, 1);
    if (target.getDay() !== 4) {
        target.setMonth(0, 1 + ((4 - target.getDay()) + 7) % 7);
    }
    return 1 + Math.ceil((firstThursday - target) / 604800000);
}

const getWeeksInYear = (year) => {
    const d = new Date(year, 11, 31);
    const week = getISOWeekNumber(d);
    if (week === 1) {
        const d2 = new Date(year, 11, 24);
        return getISOWeekNumber(d2);
    }

    return week;
}

const getYearOptions = () => {
    const startYear = 2025;
    const endYear = new Date().getFullYear();
    const min = Math.min(startYear, endYear);
    const max = Math.max(startYear, endYear);
    const years = [];
    for (let y = min; y <= max; y++) {
        years.push(y + '');
    }
    return years.sort((a, b) => b - a)
}

const getWeekOptions = () => {
    const year = new Date().getFullYear();
    const totalWeeks = getWeeksInYear(year);
    const options = [];
    for (let w = 1; w <= totalWeeks; w++) {
        const dateRange = getWeekDateRange(year, w);
        options.push({
            value: w,
            label: `Week ${w} (${dateRange})`
        });
    }
    return options;
}
const formatMySQLDate = (date) => {
    const y = date.getFullYear();
    const m = (date.getMonth() + 1).toString().padStart(2, '0');
    const d = date.getDate().toString().padStart(2, '0');
    return `${y}-${m}-${d}`;
};
export const HeatContent = (props) => {


    const [year, setYear] = useState(new Date().getFullYear());
    const [week, setWeek] = useState( getISOWeekNumber(new Date()));
    const handleYearChange = (e) => {
        setYear(e.target.value);
        setWeek(1);
    };
    const handleWeekChange = (e) => {
        const val = e.target ? parseInt(e.target.value, 10) : parseInt(e, 10);
        setWeek(val);
    };

    const years = getYearOptions();
    const weeks = getWeekOptions();

    const [chartData, setChartData] = useState([]);

    useEffect(() => {

        const simple = new Date(year, 0, 4);
        const day = simple.getDay() || 7;
        const mondayOfWeek1 = new Date(simple.getFullYear(), 0, 4 - day + 1);
        const startObj = new Date(mondayOfWeek1);
        startObj.setDate(mondayOfWeek1.getDate() + (week - 1) * 7);
        const endObj = new Date(startObj);
        endObj.setDate(startObj.getDate() + 6);


        axios.get(`/api/user/activity?uid=${props.uid}&start=${formatMySQLDate(startObj)}&end=${formatMySQLDate(endObj)}`).then((response) => {
            setChartData(response.data.data);
        })
    },[year,week])
    return (
        <div style={{ padding: 20, border: '1px solid #ccc', borderRadius: 8}}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 15 }}>
                <div className={'flex sm:flex-row flex-col'}>
                    <Select
                        value={year }
                        onChange={handleYearChange}
                        className={'max-w-xs'}
                        defaultSelectedKeys={[new Date().getFullYear() + ''] }
                    >
                        {years.map((animal) => (
                            <SelectItem key={animal}>{animal}</SelectItem>
                        ))}
                    </Select>
                    <Select
                        value={week}
                        onChange={handleWeekChange}
                        defaultSelectedKeys={[getISOWeekNumber(new Date()) + ''] }
                        className={'max-w-xs'}
                    >
                        {weeks.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value}>
                                {opt.label}
                            </SelectItem>
                        ))}
                    </Select>
                </div>
                {chartData.length > 0 && (
                    <HeatMap data={chartData}/>
                )}
            </div>
        </div>
    );
};
