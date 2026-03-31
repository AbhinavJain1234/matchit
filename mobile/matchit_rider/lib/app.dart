import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  String responseData = "";

  Future<void> createRide() async {
    print("Button Pressed");
    try {
      final response = await http.post(
        Uri.parse('http://192.168.0.101:8080/ride/request'),
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({
          "rider_id": "rider_123",
          "pickup_lat": 28.6139,
          "pickup_lon": 77.2090,
          "dest_lat": 28.6519,
          "dest_lon": 77.2315,
        }),
      );

      print(response.body.toString());

      setState(() {
        responseData = jsonDecode(response.body).toString();
      });
    } catch (e) {
      print("Error: $e");
    }
  }

  Future<void> cancelRide() async {
    print("Button Pressed");
    try {
      final response = await http.post(
        Uri.parse('http://192.168.0.101:8080/ride/cancelRequest'),
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({
          "rider_id": "rider_123",
          "ride_id": "b48c7ff1-e3d3-4244-ae73-2c9e5c15929e",
        }),
      );

      setState(() {
        responseData = jsonDecode(response.body).toString();
      });
    } catch (e) {
      print("Error: $e");
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Demo',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: Scaffold(
        appBar: AppBar(title: const Text('MatchIt Rider')),
        body: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            ElevatedButton(
              onPressed: createRide,
              child: Text("Create Ride Request"),
            ),
            Text(responseData),
            ElevatedButton(
              onPressed: cancelRide,
              child: Text("Cancel Ride Request"),
            ),
          ],
        ),
      ),
    );
  }
}
