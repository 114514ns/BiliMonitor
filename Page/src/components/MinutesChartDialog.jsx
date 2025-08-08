import React, { useEffect } from 'react';
import { Modal, ModalBody, ModalContent, ModalHeader } from "@heroui/react";
import axios from "axios";
import { LineChart } from '@mui/x-charts/LineChart';
import { BarChart } from '@mui/x-charts/BarChart';
import {Slider} from "@mui/material";

function MinutesChartDialog(props) {
    const [data, setData] = React.useState([]);
    const host = location.hostname;
    const port =  location.port;
    const protocol = location.protocol.replace(":", "");
    const [open, setOpen] = React.useState(true);
    const [value, setValue] = React.useState([0 ,25]);
    const [max,setMax] = React.useState(0);
    const minDistance = 10;
    const handleChange = (event, newValue, activeThumb) => {
        if (!Array.isArray(newValue)) {
            return;
        }
        if (newValue[1] - newValue[0] < minDistance) {
            if (activeThumb === 0) {
                const clamped = Math.min(newValue[0], 100 - minDistance);
                setValue([clamped, clamped + minDistance]);
            } else {
                const clamped = Math.max(newValue[1], minDistance);
                setValue([clamped - minDistance, clamped]);
            }
        } else {
            setValue(newValue);
        }
        console.log(newValue);
    };
    // 获取MinuteTime作为x轴
    function getKeys(arr) {
        return arr.map(obj => obj.MinuteTime).slice(value[0],value[1]);
    }

    // 获取RecordCount作为y轴
    function getValues(arr) {
        return arr.map(obj => obj.RecordCount).slice(value[0],value[1]);
    }

    useEffect(() => {
        console.log(props);
        axios.get(`${protocol}://${host}:${port}/api/chart/live?id=${props.id}`).then((response) => {
            setData(response.data.data);
            setMax(response.data.data.length)
        });
    }, [props.id]);

    return (
        <div>
            <Modal isOpen={open} onClose={() => {
                setOpen(false);
                props.onClose();
            }} size="2xl">
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">弹幕分布</ModalHeader>
                    <ModalBody>
                        <BarChart
                            xAxis={[{ scaleType: 'band', data:getKeys(data) }]}
                            series={[{ data: getValues(data) }]}
                            width={600}
                            height={300}
                        />
                        <Slider
                            value={value}
                            onChange={handleChange}
                            valueLabelDisplay="auto"
                            min={0}
                            max={max}
                            sx={{ mt: 2 }}
                        />
                    </ModalBody>
                </ModalContent>
            </Modal>
        </div>
    );
}

export default MinutesChartDialog;
