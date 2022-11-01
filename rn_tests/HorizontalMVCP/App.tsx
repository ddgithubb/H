import React, { useEffect, useState } from 'react'
import { View, Text } from 'react-native'
import ScrollView from './src/InfiniteHorizontalScrollView/ScrollView'

let data = [];
let key = -1;
let key2 = 100;
for (let i = 0; i < 100; i++) {
    data.push(i);
}

export default function App() {

    const [ d, sd ] = useState<number[]>(data);

    useEffect(() => {
        setInterval(() => {
            for (let i = 0; i < 10; i++) {
                data.unshift(key);
                key--;
            }
            // data.push(key2);
            // data.shift();
            sd([...data]);
            key2++;
        }, 1000);
    }, []);

    return (
        <>
            <Text style={{ fontSize: 50, color: "#FFFFFF" }}>TITLE</Text>
            <ScrollView
                style={{ width: "100%" }}
                maintainVisibleContentPosition={{
                    minIndexForVisible: 0
                }}
                //onScroll={({ nativeEvent }) => console.log(nativeEvent)}
                >
                {/* <Text style={{ fontSize: 20 }}>LOADING...</Text> */}
                {
                    d.map((item, index) => {
                        return <Text key={item} style={{ fontSize: 30, color: "#FFFFFF", width: 50 }}>{ item } </Text>
                    })
                }
                <Text style={{ fontSize: 10, color: "#FFFFFF" }}>Showing...</Text>
            </ScrollView>
        </>
    )
}